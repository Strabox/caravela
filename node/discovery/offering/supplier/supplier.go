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
	"github.com/strabox/caravela/util"
	"sync"
	"time"
)

// Supplier handles all the logic of managing the node own resources, advertising them into the system.
type Supplier struct {
	nodeCommon.NodeComponent // Base component

	config         *configuration.Configuration // System's configurations.
	offersStrategy OfferingStrategy             // Encapsulates the strategies to manage the offers in the system.
	client         external.Caravela            // Client to collaborate with other CARAVELA's nodes

	nodeGUID *guid.GUID

	activeOffers       map[common.OfferID]*supplierOffer // Map with the current activeOffers (that are being managed by traders)
	offersIDGen        common.OfferID                    // Monotonic counter to generate offer's local unique IDs
	offersMutex        sync.Mutex                        // Mutex to handle active offers management
	resourcesMap       *resources.Mapping                // The resources<->GUID mapping
	maxResources       *resources.Resources              // The maximum resources that the Docker engine has available (Static value)
	availableResources *resources.Resources              // CURRENT Available resources to offer

	quitChan             chan bool        // Channel to alert that the node is stopping
	supplyingTicker      <-chan time.Time // Timer to supply available resources
	refreshesCheckTicker <-chan time.Time // Timer to check if the active offers are in alive traders
}

// NewSupplier creates a new supplier component, that manages the local resources.
func NewSupplier(config *configuration.Configuration, overlay external.Overlay, client external.Caravela,
	resourcesMap *resources.Mapping, maxResources resources.Resources) *Supplier {

	s := &Supplier{
		config:         config,
		offersStrategy: CreateOffersStrategy(config),
		client:         client,

		nodeGUID: nil,

		resourcesMap:       resourcesMap,
		maxResources:       maxResources.Copy(),
		availableResources: maxResources.Copy(),
		offersIDGen:        0,
		activeOffers:       make(map[common.OfferID]*supplierOffer),
		offersMutex:        sync.Mutex{},

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

	if s.nodeGUID != nil {
		ctx = context.WithValue(ctx, types.NodeGUIDKey, s.nodeGUID.String())
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
func (s *Supplier) ObtainResources(offerID int64, resourcesNecessary resources.Resources) bool {
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
func (s *Supplier) ReturnResources(releasedResources resources.Resources) {
	if !s.IsWorking() {
		panic(errors.New("can't return resources, supplier not working"))
	}

	s.offersMutex.Lock()
	defer s.offersMutex.Unlock()

	log.Debugf(util.LogTag("SUPPLIER")+"RESOURCES RELEASED Res: <%d;%d>", releasedResources.CPUs(), releasedResources.RAM())
	s.availableResources.Add(releasedResources)

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
		s.offersStrategy.UpdateOffers(*s.availableResources)
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
			s.availableResources.CPUs(), s.availableResources.RAM(), s.maxResources.CPUs(), s.maxResources.RAM()))
	}
	if !s.maxResources.Contains(*s.availableResources) {
		panic(fmt.Errorf("there are more resources than the maximum, available: <%d,%d>, max: <%d,%d>",
			s.availableResources.CPUs(), s.availableResources.RAM(), s.maxResources.CPUs(), s.maxResources.RAM()))
	}
}

// Simulation
func (s *Supplier) SetNodeGUID(GUID guid.GUID) {
	if s.nodeGUID == nil {
		s.nodeGUID = guid.NewGUIDBytes(GUID.Bytes())
	}
}

// Simulation
func (s *Supplier) AvailableResources() types.Resources {
	s.offersMutex.Lock()
	defer s.offersMutex.Unlock()

	return types.Resources{
		CPUClass: types.CPUClass(s.availableResources.CPUClass()),
		CPUs:     s.availableResources.CPUs(),
		RAM:      s.availableResources.RAM(),
	}
}

// Simulation
func (s *Supplier) MaximumResources() types.Resources {
	return types.Resources{
		CPUClass: types.CPUClass(s.availableResources.CPUClass()),
		CPUs:     s.maxResources.CPUs(),
		RAM:      s.maxResources.RAM(),
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
