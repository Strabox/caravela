package supplier

import (
	"context"
	"errors"
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
	offersStrategy OffersManager                // Encapsulates the strategies to manage the offers in the system.
	client         external.Caravela            // Client to collaborate with other CARAVELA's nodes

	nodeGUID *guid.GUID

	resourcesMap       *resources.Mapping                // The resources<->GUID mapping
	maxResources       *resources.Resources              // The maximum resources that the Docker engine has available (Static value)
	availableResources *resources.Resources              // CURRENT Available resources to offer
	offersIDGen        common.OfferID                    // Monotonic counter to generate offer's local unique IDs
	activeOffers       map[common.OfferID]*supplierOffer // Map with the current activeOffers (that are being managed by traders)
	offersMutex        *sync.Mutex                       // Mutex to handle active offers management

	quitChan             chan bool        // Channel to alert that the node is stopping
	supplyingTicker      <-chan time.Time // Timer to supply available resources
	refreshesCheckTicker <-chan time.Time // Timer to check if the active offers are in alive traders
}

// NewSupplier creates a new supplier component, that manages the local resources.
func NewSupplier(config *configuration.Configuration, overlay external.Overlay, client external.Caravela,
	resourcesMap *resources.Mapping, maxResources resources.Resources) *Supplier {

	offersStrategy := CreateOffersStrategy(config)
	offersStrategy.Init(resourcesMap, overlay, client)
	return &Supplier{
		config:         config,
		offersStrategy: offersStrategy,
		client:         client,

		nodeGUID: nil,

		resourcesMap:       resourcesMap,
		maxResources:       maxResources.Copy(),
		availableResources: maxResources.Copy(),
		offersIDGen:        0,
		activeOffers:       make(map[common.OfferID]*supplierOffer),
		offersMutex:        &sync.Mutex{},

		quitChan:             make(chan bool),
		supplyingTicker:      time.NewTicker(config.SupplyingInterval()).C,
		refreshesCheckTicker: time.NewTicker(config.RefreshesCheckInterval()).C,
	}
}

// startSupplying controls the time dependant actions like supplying the resources.
func (sup *Supplier) startSupplying() {
	for {
		select {
		case <-sup.supplyingTicker: // Offer the available resources into a random trader (responsible for them).
			go func() {
				sup.offersMutex.Lock()
				defer sup.offersMutex.Unlock()
				sup.createOffer()
			}()
		case <-sup.refreshesCheckTicker: // Check if the activeOffers are being refreshed by the respective trader
			sup.offersMutex.Lock()

			for offerKey, offer := range sup.activeOffers {
				offer.VerifyRefreshes(sup.config.RefreshMissedTimeout())

				if offer.RefreshesMissed() >= sup.config.MaxRefreshesMissed() {
					log.Debugf(util.LogTag("SUPPLIER")+"Offer DOWN, Offer: %d, HandlerTrader: %s",
						offer.ID(), offer.ResponsibleTraderIP())

					sup.availableResources.Add(*offer.Resources())
					delete(sup.activeOffers, offerKey)
				}
			}

			sup.offersMutex.Unlock()
		case res := <-sup.quitChan: // Stopping the supplier
			if res {
				log.Infof(util.LogTag("SUPPLIER") + "STOPPED")
				return
			}
		}
	}
}

// Find a list active Offers that best suit the target resources given.
func (sup *Supplier) FindOffers(ctx context.Context, targetResources resources.Resources) []types.AvailableOffer {
	if !sup.IsWorking() {
		panic(errors.New("can't find offers, supplier not working"))
	}

	if !targetResources.IsValid() { // If the resource combination is not valid we will search for the lowest one
		targetResources = *sup.resourcesMap.LowestResources()
	}

	if sup.nodeGUID != nil {
		ctx = context.WithValue(ctx, types.NodeGUIDKey, sup.nodeGUID.String())
	}
	return sup.offersStrategy.FindOffers(ctx, targetResources)
}

// Tries refresh an offer. Called when a refresh message was received.
func (sup *Supplier) RefreshOffer(fromTrader *types.Node, refreshOffer *types.Offer) bool {
	if !sup.IsWorking() {
		panic(errors.New("can't refresh offer, supplier not working"))
	}

	sup.offersMutex.Lock()
	defer sup.offersMutex.Unlock()

	offer, exist := sup.activeOffers[common.OfferID(refreshOffer.ID)]

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
func (sup *Supplier) ObtainResources(offerID int64, resourcesNecessary resources.Resources) bool {
	if !sup.IsWorking() {
		panic(errors.New("can't obtain resources, supplier not working"))
	}

	sup.offersMutex.Lock()
	defer sup.offersMutex.Unlock()

	supOffer, exist := sup.activeOffers[common.OfferID(offerID)]
	if !exist || !supOffer.Resources().Contains(resourcesNecessary) || !sup.availableResources.Contains(resourcesNecessary) { // Offer does not exist in the supplier OR asking more resources than the offer has available
		return false
	} else {
		sup.availableResources.Sub(resourcesNecessary)

		delete(sup.activeOffers, common.OfferID(offerID))

		removeOffer := func() {
			sup.client.RemoveOffer(
				context.Background(),
				&types.Node{IP: sup.config.HostIP(), GUID: ""},
				&types.Node{IP: supOffer.ResponsibleTraderIP(), GUID: supOffer.ResponsibleTraderGUID().String()},
				&types.Offer{ID: int64(supOffer.ID())},
			)
		}

		if sup.config.Simulation() {
			removeOffer()
			sup.createOffer() // Update its own offers
		} else {
			go removeOffer()
			go func() {
				sup.offersMutex.Lock()
				defer sup.offersMutex.Unlock()
				sup.createOffer()
			}() // Update its own offers
		}
		return true
	}
}

// Release resources of an used offer into the supplier again in order to offer them again into the system.
func (sup *Supplier) ReturnResources(releasedResources resources.Resources) {
	if !sup.IsWorking() {
		panic(errors.New("can't return resources, supplier not working"))
	}

	sup.offersMutex.Lock()
	defer sup.offersMutex.Unlock()

	sup.availableResources.Add(releasedResources)

	if sup.config.Simulation() {
		sup.createOffer() // Update its own offers sequential
	} else {
		go func() {
			sup.offersMutex.Lock()
			defer sup.offersMutex.Unlock()
			sup.createOffer()
		}() // Update its own offers in the background
	}
}

func (sup *Supplier) createOffer() {
	sup.checkResourcesInvariant() // Runtime resources assertion!!!
	if sup.availableResources.IsValid() {
		lowerPartitions, _ := sup.resourcesMap.LowerPartitionsOffer(*sup.availableResources)
		offersToRemove := make([]*supplierOffer, 0)

	OfferLoop:
		for _, offer := range sup.activeOffers {
			offerPartition := sup.resourcesMap.ResourcesByGUID(*offer.ResponsibleTraderGUID())
			for lp, lowerPartition := range lowerPartitions {
				if offerPartition.Equals(lowerPartition) {
					lowerPartitions = append(lowerPartitions[:lp], lowerPartitions[lp+1:]...)
					continue OfferLoop
				}
			}
			offersToRemove = append(offersToRemove, offer)
		}

		for _, offerToRemove := range offersToRemove {
			sup.client.RemoveOffer(
				context.Background(),
				&types.Node{IP: sup.config.HostIP(), GUID: ""},
				&types.Node{IP: offerToRemove.ResponsibleTraderIP(), GUID: offerToRemove.ResponsibleTraderGUID().String()},
				&types.Offer{ID: int64(offerToRemove.ID())})
		}

		for _, toOffer := range lowerPartitions {
			offer, err := sup.offersStrategy.CreateOffer(int64(sup.offersIDGen), toOffer)
			if err == nil {
				sup.activeOffers[offer.ID()] = offer
				sup.availableResources.SetZero()
			}
			sup.offersIDGen++
		}
	}
}

func (sup *Supplier) checkResourcesInvariant() {
	if sup.availableResources.IsNegative() {
		panic(errors.New("there are negative resources available :|"))
	}
	if !sup.maxResources.Contains(*sup.availableResources) {
		panic(errors.New("there are more resources than the maximum available :|"))
	}
}

// Simulation
func (sup *Supplier) SetNodeGUID(GUID guid.GUID) {
	if sup.nodeGUID == nil {
		sup.nodeGUID = guid.NewGUIDBytes(GUID.Bytes())
	}
}

// Simulation
func (sup *Supplier) AvailableResources() types.Resources {
	sup.offersMutex.Lock()
	defer sup.offersMutex.Unlock()

	return types.Resources{
		CPUs: sup.availableResources.CPUs(),
		RAM:  sup.availableResources.RAM(),
	}
}

// Simulation
func (sup *Supplier) MaximumResources() types.Resources {
	return types.Resources{
		CPUs: sup.maxResources.CPUs(),
		RAM:  sup.maxResources.RAM(),
	}
}

// ===============================================================================
// =							SubComponent Interface                           =
// ===============================================================================

func (sup *Supplier) Start() {
	sup.Started(sup.config.Simulation(), func() {
		if !sup.config.Simulation() {
			go sup.startSupplying()
		} else {
			sup.offersMutex.Lock()
			defer sup.offersMutex.Unlock()
			sup.createOffer()
		}
	})
}

func (sup *Supplier) Stop() {
	sup.Stopped(func() {
		sup.quitChan <- true
	})
}

func (sup *Supplier) IsWorking() bool {
	return sup.Working()
}
