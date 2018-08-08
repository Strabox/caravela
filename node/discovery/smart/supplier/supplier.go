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

	// Offer does not exist in the supplier OR asking more resources than the offer has available
	if !exist || !supOffer.Resources().Contains(resourcesNecessary) {
		return false
	} else {
		remainingResources := supOffer.Resources().Copy()
		remainingResources.Sub(resourcesNecessary)

		sup.availableResources.Add(*remainingResources)

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
	if sup.availableResources.IsValid() {
		// What?: Remove all active offers from the traders in order to gather all available resources.
		// Goal: This is used to try offer the maximum amount of resources the node has available between
		//		 the Available (offered) and the Available (but not offered).
		for offerID, offer := range sup.activeOffers {
			removeOffer := func(offer *supplierOffer) {
				sup.client.RemoveOffer(
					context.Background(),
					&types.Node{IP: sup.config.HostIP(), GUID: ""},
					&types.Node{IP: offer.ResponsibleTraderIP(), GUID: offer.ResponsibleTraderGUID().String()},
					&types.Offer{ID: int64(offer.ID())},
				)
			}
			if sup.config.Simulation() {
				removeOffer(offer) // Send remove offer message sequential
			} else {
				go removeOffer(offer) // Send remove offer message in background
			}

			delete(sup.activeOffers, offerID)
			sup.availableResources.Add(*offer.Resources())
		}

		log.Debugf(util.LogTag("SUPPLIER")+"CREATING offer... Offer: %d, Res: <%d;%d>",
			int64(sup.offersIDGen), sup.availableResources.CPUs(), sup.availableResources.RAM())

		offer, err := sup.offersStrategy.CreateOffer(int64(sup.offersIDGen), *sup.availableResources)
		if err == nil {
			sup.activeOffers[offer.ID()] = offer
			sup.availableResources.SetZero()
		}
		sup.offersIDGen++
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

	res := types.Resources{
		CPUs: sup.availableResources.CPUs(),
		RAM:  sup.availableResources.RAM(),
	}

	for _, offer := range sup.activeOffers {
		res.CPUs += offer.Resources().CPUs()
		res.RAM += offer.Resources().RAM()
	}

	return res
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
