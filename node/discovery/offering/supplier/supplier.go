package supplier

import (
	"context"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	nodeCommon "github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/discovery/common"
	"github.com/strabox/caravela/node/external"
	"github.com/strabox/caravela/overlay"
	"github.com/strabox/caravela/util"
	"github.com/strabox/caravela/util/debug"
	"sync"
	"time"
	"unsafe"
)

// Supplier handles all the logic of managing the node own resources, advertising them into the system.
type Supplier struct {
	nodeCommon.NodeComponent // Base component

	config         *configuration.Configuration // System's configurations.
	offersStrategy OfferingStrategy             // Encapsulates the strategies to manage the offers in the system.
	client         external.Caravela            // Client to collaborate with other CARAVELA's nodes

	activeOffers       map[common.OfferID]*supplierOffer // Map with the current activeOffers (that are being managed by traders)
	offersIDGen        common.OfferID                    // Monotonic counter to generate offer's local unique IDs
	offersMutex        sync.Mutex                        // Mutex to handle active offers management
	resourcesMap       *resources.Mapping                // The resources<->GUID mapping
	maxResources       *resources.Resources              // The maximum resources that the Docker engine has available (Static value)
	availableResources *resources.Resources              // CURRENT Available resources to offer
	containersRunning  int                               // Number of containers running in the node.

	quitChan             chan bool        // Channel to alert that the node is stopping
	supplyingTicker      <-chan time.Time // Timer to supply available resources
	refreshesCheckTicker <-chan time.Time // Timer to check if the active offers are in alive traders
}

// NewSupplier creates a new supplier component, that manages the local resources.
func NewSupplier(node nodeCommon.Node, config *configuration.Configuration, overlay overlay.Overlay, client external.Caravela,
	resourcesMap *resources.Mapping, maxResources resources.Resources) *Supplier {

	s := &Supplier{
		config:         config,
		offersStrategy: CreateOffersStrategy(node, config),
		client:         client,

		resourcesMap:       resourcesMap,
		maxResources:       maxResources.Copy(),
		availableResources: maxResources.Copy(),
		offersIDGen:        0,
		activeOffers:       make(map[common.OfferID]*supplierOffer),
		offersMutex:        sync.Mutex{},
		containersRunning:  0,

		quitChan:             make(chan bool),
		supplyingTicker:      time.NewTicker(config.SupplyingInterval()).C,
		refreshesCheckTicker: time.NewTicker(config.RefreshesCheckInterval()).C,
	}
	s.offersStrategy.Init(s, resourcesMap, overlay, client)
	return s
}

// start controls the time dependant actions like supplying the resources.
func (s *Supplier) start() {
	for {
		select {
		case <-s.supplyingTicker: // Offer the available resources into a random trader (responsible for them).
			go func() {
				s.offersMutex.Lock()
				defer s.offersMutex.Unlock()
				s.updateOffers()
			}()
		case <-s.refreshesCheckTicker: // Check if the activeOffers are being refreshed by the respective trader
			s.offersMutex.Lock()

			offerDown := false
			for offerKey, offer := range s.activeOffers {
				offer.VerifyRefreshes(s.config.RefreshMissedTimeout())

				if offer.RefreshesMissed() >= s.config.MaxRefreshesMissed() {
					log.Debugf(util.LogTag("SUPPLIER")+"Offer DOWN, Offer: %d, HandlerTrader: %s",
						offer.ID(), offer.ResponsibleTraderIP())
					offerDown = true
					delete(s.activeOffers, offerKey)
				}
			}

			if offerDown {
				s.updateOffers()
			}

			s.offersMutex.Unlock()
		case res := <-s.quitChan: // Stopping the supplier
			if res {
				log.Infof(util.LogTag("SUPPLIER") + "STOPPED")
				return
			}
		}
	}
}

// Find a list active Offers that best suit the target resources given.
func (s *Supplier) FindOffers(ctx context.Context, targetResources resources.Resources) []types.AvailableOffer {
	if !s.IsWorking() {
		panic(errors.New("can't find offers, supplier not working"))
	}

	if !targetResources.IsValid() { // If the resource combination is not valid we will search for the lowest one
		targetResources = *s.resourcesMap.LowestResources()
	}

	return s.offersStrategy.FindOffers(ctx, targetResources)
}

// Tries refresh an offer. Called when a refresh message was received.
func (s *Supplier) RefreshOffer(fromTrader *types.Node, refreshOffer *types.Offer) bool {
	if !s.IsWorking() {
		panic(errors.New("can't refresh offer, supplier not working"))
	}

	s.offersMutex.Lock()
	defer s.offersMutex.Unlock()

	offer, exist := s.activeOffers[common.OfferID(refreshOffer.ID)]

	if !exist {
		log.Debugf(util.LogTag("SUPPLIER")+"Offer: %d refresh FAILED (Offer does not exist)", refreshOffer.ID)
		return false
	}

	if offer.IsResponsibleTrader(*guid.NewGUIDString(fromTrader.GUID)) {
		offer.Refresh()
		log.Debugf(util.LogTag("SUPPLIER")+"Offer: %d refresh SUCCESS", refreshOffer.ID)
		return true
	} else {
		log.Debugf(util.LogTag("SUPPLIER")+"Offer: %d refresh FAILED (wrong trader)", refreshOffer.ID)
		return false
	}
}

// Tries to obtain a subset of the resources represented by the given offer in order to deploy  a container.
// It updates the respective trader that manages the offer.
func (s *Supplier) ObtainResources(offerID int64, resourcesNecessary resources.Resources, numContainersToRun int) bool {
	if !s.IsWorking() {
		panic(errors.New("can't obtain resources, supplier not working"))
	}

	s.offersMutex.Lock()
	defer s.offersMutex.Unlock()

	supOffer, exist := s.activeOffers[common.OfferID(offerID)]
	if !exist || !supOffer.Resources().Contains(resourcesNecessary) || !s.availableResources.Contains(resourcesNecessary) { // Offer does not exist in the supplier OR asking more resources than the offer has available
		return false
	} else {
		s.availableResources.Sub(resourcesNecessary)
		s.containersRunning += numContainersToRun

		delete(s.activeOffers, common.OfferID(offerID))

		removeOffer := func() {
			s.client.RemoveOffer(
				context.Background(),
				&types.Node{IP: s.config.HostIP()},
				&types.Node{IP: supOffer.ResponsibleTraderIP(), GUID: supOffer.ResponsibleTraderGUID().String()},
				&types.Offer{ID: int64(supOffer.ID())},
			)
		}

		if s.config.Simulation() {
			removeOffer()
			s.updateOffers() // Update its own offers
		} else {
			go removeOffer()
			go func() {
				s.offersMutex.Lock()
				defer s.offersMutex.Unlock()
				s.updateOffers()
			}() // Update its own offers
		}
		return true
	}
}

// Release resources of an used offer into the supplier again in order to offer them again into the system.
func (s *Supplier) ReturnResources(releasedResources resources.Resources, numContainersStopped int) {
	if !s.IsWorking() {
		panic(errors.New("can't return resources, supplier not working"))
	}

	s.offersMutex.Lock()
	defer s.offersMutex.Unlock()

	log.Debugf(util.LogTag("SUPPLIER")+"RESOURCES RELEASED Res: <%d;%d>", releasedResources.CPUs(), releasedResources.Memory())
	s.availableResources.Add(releasedResources)
	s.containersRunning -= numContainersStopped

	if s.config.Simulation() {
		s.updateOffers() // Update its own offers sequential
	} else {
		go func() {
			s.offersMutex.Lock()
			defer s.offersMutex.Unlock()
			s.updateOffers()
		}() // Update its own offers in the background
	}
}

func (s *Supplier) updateOffers() {
	s.checkResourcesInvariant() // Runtime resources assertion!!!
	if s.availableResources.IsValid() {
		usedResources := s.maxResources.Copy()
		usedResources.Sub(*s.availableResources)

		s.offersStrategy.UpdateOffers(context.Background(), *s.availableResources, *usedResources)
	}
}

func (s *Supplier) forceOfferRefresh(offerID common.OfferID, success bool) {
	if offer, exist := s.activeOffers[offerID]; exist {
		if success {
			offer.Refresh()
		} else {
			offer.UnreachableTrader()
		}
	}
}

func (s *Supplier) newOfferID() common.OfferID {
	res := s.offersIDGen
	s.offersIDGen++
	return res
}

func (s *Supplier) addOffer(offer *supplierOffer) {
	s.activeOffers[offer.ID()] = offer
}

func (s *Supplier) removeOffer(offerID common.OfferID) {
	delete(s.activeOffers, offerID)
}

func (s *Supplier) numContainersRunning() int {
	return s.containersRunning
}

func (s *Supplier) offers() []supplierOffer {
	res := make([]supplierOffer, len(s.activeOffers))
	i := 0
	for _, supOffer := range s.activeOffers {
		res[i] = *supOffer
		i++
	}
	return res
}

func (s *Supplier) checkResourcesInvariant() {
	if s.availableResources.IsNegative() {
		panic(fmt.Errorf("more resources being used than maximum, available: <%d,%d>, max: <%d,%d>",
			s.availableResources.CPUs(), s.availableResources.Memory(), s.maxResources.CPUs(), s.maxResources.Memory()))
	}
	if !s.maxResources.Contains(*s.availableResources) {
		panic(fmt.Errorf("there are more resources than the maximum, available: <%d,%d>, max: <%d,%d>",
			s.availableResources.CPUs(), s.availableResources.Memory(), s.maxResources.CPUs(), s.maxResources.Memory()))
	}
}

//Simulation
func (s *Supplier) NumActiveOffers() int {
	s.offersMutex.Lock()
	defer s.offersMutex.Unlock()
	return len(s.activeOffers)
}

// Simulation
func (s *Supplier) AvailableResources() types.Resources {
	s.offersMutex.Lock()
	defer s.offersMutex.Unlock()

	return types.Resources{
		CPUClass: types.CPUClass(s.availableResources.CPUClass()),
		CPUs:     s.availableResources.CPUs(),
		Memory:   s.availableResources.Memory(),
	}
}

// Simulation
func (s *Supplier) MaximumResources() types.Resources {
	return types.Resources{
		CPUClass: types.CPUClass(s.maxResources.CPUClass()),
		CPUs:     s.maxResources.CPUs(),
		Memory:   s.maxResources.Memory(),
	}
}

// ===============================================================================
// =							SubComponent Interface                           =
// ===============================================================================

func (s *Supplier) Start() {
	s.Started(s.config.Simulation(), func() {
		if !s.config.Simulation() {
			go s.start()
		} else {
			s.offersMutex.Lock()
			defer s.offersMutex.Unlock()
			s.updateOffers()
		}
	})
}

func (s *Supplier) Stop() {
	s.Stopped(func() {
		s.quitChan <- true
	})
}

func (s *Supplier) IsWorking() bool {
	return s.Working()
}

// ===============================================================================
// =							    Debug Methods                                =
// ===============================================================================

func (s *Supplier) DebugSizeBytes() int {
	supplierOfferSizeBytes := func(offer *supplierOffer) uintptr {
		offerSizeBytes := unsafe.Sizeof(*offer)
		// common.Offer
		offerSizeBytes += unsafe.Sizeof(*offer.Offer)
		offerSizeBytes += debug.DebugSizeofResources(offer.Offer.Resources())
		// supplier offer
		offerSizeBytes += debug.DebugSizeofString(offer.responsibleTraderIP)
		offerSizeBytes += debug.DebugSizeofGUID(offer.responsibleTraderGUID)
		return offerSizeBytes
	}

	supplierSizeBytes := unsafe.Sizeof(*s)
	supplierSizeBytes += debug.DebugSizeofResources(s.maxResources)
	supplierSizeBytes += debug.DebugSizeofResources(s.availableResources)
	supplierSizeBytes += 30 // Hack: Offer strategy structure.
	for offerID, offer := range s.activeOffers {
		supplierSizeBytes += unsafe.Sizeof(offerID)
		supplierSizeBytes += unsafe.Sizeof(offer)
		supplierSizeBytes += supplierOfferSizeBytes(offer)
	}
	return int(supplierSizeBytes)
}
