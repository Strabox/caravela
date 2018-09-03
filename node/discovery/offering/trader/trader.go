package trader

import (
	"context"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	nodeCommon "github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/discovery/common"
	"github.com/strabox/caravela/node/external"
	overlayTypes "github.com/strabox/caravela/overlay/types"
	"github.com/strabox/caravela/util"
	"sync"
	"time"
)

// Trader is responsible for managing offers from multiple suppliers and negotiate these offers with buyers.
// The resources combination that the trader can handle is described by its GUID.
type Trader struct {
	nodeCommon.NodeComponent // Base component

	config  *configuration.Configuration // System's configurations
	overlay external.Overlay             // Node overlay to efficient route messages to specific nodes.
	client  external.Caravela            // Client for the system

	guid             *guid.GUID           // Trader's own GUID
	resourcesMap     *resources.Mapping   // GUID<->FreeResources mapping
	handledResources *resources.Resources // Combination of resources that its responsible for managing (FIXED)

	nearbyTradersOffering *nearbyTradersOffering    // Nearby traders that might have offers available
	offers                map[offerKey]*traderOffer // Map with all the offers that the trader is managing
	offersMutex           *sync.Mutex               // Mutex for managing the offers

	quitChan            chan bool        // Channel to alert that the node is stopping
	refreshOffersTicker <-chan time.Time // Time ticker for sending the refreshing offer messages
	spreadOffersTimer   <-chan time.Time // Time ticker to spread offer information into the neighbors
}

// NewTrader creates a new "virtual" trader.
func NewTrader(config *configuration.Configuration, overlay external.Overlay, client external.Caravela,
	guid guid.GUID, resourcesMapping *resources.Mapping) *Trader {

	handledResources := resourcesMapping.ResourcesByGUID(guid)

	return &Trader{
		config:           config,
		overlay:          overlay,
		client:           client,
		guid:             &guid,
		resourcesMap:     resourcesMapping,
		handledResources: handledResources,

		nearbyTradersOffering: newNeighborTradersOffering(),
		offers:                make(map[offerKey]*traderOffer),
		offersMutex:           &sync.Mutex{},

		quitChan:            make(chan bool),
		refreshOffersTicker: time.NewTicker(config.RefreshingInterval()).C,
		spreadOffersTimer:   time.NewTicker(config.SpreadOffersInterval()).C,
	}
}

// start runs a endless loop goroutine that dispatches timer events into other goroutines.
func (t *Trader) start() {
	for {
		select {
		case <-t.refreshOffersTicker: // Time to refresh all the current offers (verify if the suppliers are alive)
			t.offersMutex.Lock()

			for _, offer := range t.offers {
				if offer.Refresh() {
					go func(offer *traderOffer) {
						refreshed, err := t.client.RefreshOffer(
							context.Background(),
							&types.Node{GUID: t.guid.String()},
							&types.Node{IP: offer.supplierIP},
							&types.Offer{ID: int64(offer.ID())})

						t.offersMutex.Lock()
						defer t.offersMutex.Unlock()

						offerKEY := offerKey{supplierIP: offer.supplierIP, id: common.OfferID(offer.ID())}
						offer, exist := t.offers[offerKEY]

						if err == nil && refreshed && exist { // Offer exist and was refreshed
							log.Debugf(util.LogTag("TRADER")+"Refresh SUCCEEDED ,supplier: %s, offer: %d",
								offer.SupplierIP(), offer.ID())
							offer.RefreshSucceeded()
						} else if err == nil && !refreshed && exist { // Offer did not exist, so it was not refreshed
							log.Debugf(util.LogTag("TRADER")+"Refresh FAILED (offer did not exist),"+
								" supplier: %s, offer: %d", offer.SupplierIP(), offer.ID())
							delete(t.offers, offerKEY)
						} else if err != nil && exist { // Offer exist but the refresh message failed
							log.Debugf(util.LogTag("TRADER")+"Refresh FAILED, supplier: %s, offer: %d",
								offer.SupplierIP(), offer.ID())
							offer.RefreshFailed()
							if offer.RefreshesFailed() >= t.config.MaxRefreshesFailed() {
								log.Debugf(util.LogTag("TRADER")+"REMOVING offer, supplier: %s, offer: %d",
									offer.SupplierIP(), offer.ID())
								delete(t.offers, offerKEY)
							}
						}
					}(offer)
				}
			}

			t.offersMutex.Unlock()
		case <-t.spreadOffersTimer:
			// Advertise offers (if any) into the neighbors traders.
			// Necessary only to overcame the problems of unnoticed death of a neighbor.
			if t.haveOffers() && t.config.SpreadOffers() {
				t.advertiseOffersToNeighbors(func(neighborGUID *guid.GUID) bool { return true },
					&types.Node{GUID: t.guid.String(), IP: t.config.HostIP()})
			}
		case quit := <-t.quitChan: // Stopping the trader (returning the goroutine)
			if quit {
				log.Infof(util.LogTag("TRADER")+"Trader %s STOPPED", t.guid.Short())
				return
			}
		}
	}
}

// Receives a resource offer from other node (supplier) of the system
func (t *Trader) CreateOffer(fromSupp *types.Node, newOffer *types.Offer) {
	resourcesOffered := resources.NewResourcesCPUClass(int(newOffer.FreeResources.CPUClass), newOffer.FreeResources.CPUs, newOffer.FreeResources.RAM)

	// Verify if the offer contains the resources of trader.
	// Basically verify if the offer is bigger than the handled resources of the trader.
	if resourcesOffered.Contains(*t.handledResources) {
		t.offersMutex.Lock()

		offerKey := offerKey{supplierIP: fromSupp.IP, id: common.OfferID(newOffer.ID)}
		offer := newTraderOffer(*guid.NewGUIDString(fromSupp.GUID), fromSupp.IP, common.OfferID(newOffer.ID),
			newOffer.Amount, *resourcesOffered)

		advertise := len(t.offers) == 0

		t.offers[offerKey] = offer
		log.Debugf(util.LogTag("TRADER")+"%s Offer CREATED %dX<%d;%d>, From: %s, Offer: %d",
			t.guid.Short(), newOffer.Amount, newOffer.FreeResources.CPUs, newOffer.FreeResources.RAM,
			fromSupp.IP, newOffer.ID)

		t.offersMutex.Unlock()

		if advertise && t.config.SpreadOffers() { // If node had no offers, advertise it has now for all the neighbors
			if t.config.Simulation() {
				t.advertiseOffersToNeighbors(func(neighborGUID *guid.GUID) bool { return true },
					&types.Node{GUID: t.guid.String(), IP: t.config.HostIP()})
			} else {
				go t.advertiseOffersToNeighbors(func(neighborGUID *guid.GUID) bool { return true },
					&types.Node{GUID: t.guid.String(), IP: t.config.HostIP()})
			}
		}
	}
}

func (t *Trader) UpdateOffer(fromSupp *types.Node, offer *types.Offer) {
	t.offersMutex.Lock()
	defer t.offersMutex.Unlock()

	if traderOffer, exist := t.offers[offerKey{id: common.OfferID(offer.ID), supplierIP: fromSupp.IP}]; exist {
		newOfferRes := *resources.NewResourcesCPUClass(int(offer.FreeResources.CPUClass), offer.FreeResources.CPUs, offer.FreeResources.RAM)
		traderOffer.UpdateResources(newOfferRes, offer.Amount)
	}
}

// Returns all the offers that the trader is managing.
func (t *Trader) GetOffers(ctx context.Context, _ *types.Node, relay bool) []types.AvailableOffer {
	if t.haveOffers() || !relay || !t.config.SpreadOffers() { // Trader has offers so return them immediately or we are not relaying
		t.offersMutex.Lock()
		defer t.offersMutex.Unlock()

		availableOffers := len(t.offers)
		allOffers := make([]types.AvailableOffer, availableOffers)
		index := 0
		for _, traderOffer := range t.offers {
			allOffers[index].SupplierIP = traderOffer.SupplierIP()
			allOffers[index].ID = int64(traderOffer.ID())
			allOffers[index].Amount = traderOffer.Amount()
			allOffers[index].FreeResources = types.Resources{
				CPUClass: types.CPUClass(traderOffer.Resources().CPUClass()),
				CPUs:     traderOffer.Resources().CPUs(),
				RAM:      traderOffer.Resources().RAM(),
			}
			index++
		}
		return allOffers
	} else { // Ask for offers in the nearby neighbors that we think they have offers (via offer advertise relaying)
		resOffers := make([]types.AvailableOffer, 0)

		// Ask the successor (higher GUID)
		if successor := t.nearbyTradersOffering.Successor(); successor != nil {
			successorResourcesHandled := t.resourcesMap.ResourcesByGUID(*successor.GUID())
			if t.handledResources.Equals(*successorResourcesHandled) {
				offers, err := t.client.GetOffers( // Sends the get offers message
					ctx,
					&types.Node{GUID: t.guid.String()},
					&types.Node{IP: successor.IP(), GUID: successor.GUID().String()},
					false)
				if err == nil && len(offers) != 0 {
					resOffers = append(resOffers, offers...)
				} else if err == nil && len(offers) == 0 {
					t.nearbyTradersOffering.SetSuccessor(nil)
				}
			}

		}

		// Ask the predecessor (lower GUID)
		if predecessor := t.nearbyTradersOffering.Predecessor(); predecessor != nil {
			predecessorResourcesHandled := t.resourcesMap.ResourcesByGUID(*predecessor.GUID())
			if t.handledResources.Equals(*predecessorResourcesHandled) {
				offers, err := t.client.GetOffers( // Sends the get offers message
					ctx,
					&types.Node{GUID: t.guid.String()},
					&types.Node{IP: predecessor.IP(), GUID: predecessor.GUID().String()},
					false)
				if err == nil && len(offers) != 0 {
					resOffers = append(resOffers, offers...)
				} else if err == nil && len(offers) == 0 {
					t.nearbyTradersOffering.SetPredecessor(nil)
				}
			}

		}
		// TRY: OPTIONAl make the calls in parallel (2 goroutines) and wait here for both, then join the results.
		return resOffers
	}
}

// Remove an offer from the offering table.
func (t *Trader) RemoveOffer(fromSupp *types.Node, offer *types.Offer) {
	t.offersMutex.Lock()
	defer t.offersMutex.Unlock()

	delete(t.offers, offerKey{supplierIP: fromSupp.IP, id: common.OfferID(offer.ID)})

	log.Debugf(util.LogTag("TRADER")+"Offer REMOVED Offer: %d, Supp: %s", offer.ID, fromSupp.IP)
}

// Relay the offering advertise for the overlay neighbors if the trader doesn't have any available offers
func (t *Trader) AdvertiseNeighborOffer(fromTrader, traderOffering *types.Node) {
	fromTraderGUID := guid.NewGUIDString(fromTrader.GUID)
	traderOfferingGUID := guid.NewGUIDString(traderOffering.GUID)

	var isValidNeighbor func(neighborGUID *guid.GUID) bool = nil
	if t.guid.Higher(*fromTraderGUID) { // Message comes from a lower GUID's node
		t.nearbyTradersOffering.SetPredecessor(nodeCommon.NewRemoteNode(traderOffering.IP, *traderOfferingGUID))
		// Relay the advertise to a higher GUID's node (inside partition)
		isValidNeighbor = func(neighborGUID *guid.GUID) bool {
			return neighborGUID.Higher(*t.guid)
		}
	} else { // Message comes from a higher GUID's node
		t.nearbyTradersOffering.SetSuccessor(nodeCommon.NewRemoteNode(traderOffering.IP, *traderOfferingGUID))
		// Relay the advertise to a lower GUID's node (inside partition)
		isValidNeighbor = func(neighborGUID *guid.GUID) bool {
			return neighborGUID.Lower(*t.guid)
		}
	}

	// Do not relay the advertise if the node has offers.
	if !t.haveOffers() && t.config.SpreadOffers() {
		if t.config.Simulation() {
			t.advertiseOffersToNeighbors(isValidNeighbor, traderOffering)
		} else {
			go t.advertiseOffersToNeighbors(isValidNeighbor, traderOffering)
		}
	}
}

func (t *Trader) haveOffers() bool {
	t.offersMutex.Lock()
	defer t.offersMutex.Unlock()

	return len(t.offers) != 0
}

// Advertise that we have offers into all trader's neighbors that survive the given predicate application.
func (t *Trader) advertiseOffersToNeighbors(isValidNeighbor func(neighborGUID *guid.GUID) bool, traderOffering *types.Node) {
	log.Debugf(util.LogTag("TRADER")+"ADVERTISE offers, From: %s", t.guid.Short())

	overlayNeighbors, err := t.overlay.Neighbors(context.Background(), t.guid.Bytes())
	if err != nil {
		return
	}

	for _, overlayNeighbor := range overlayNeighbors { // Advertise to all neighbors (inside resource partition)
		advertise := func(overlayNeighbor *overlayTypes.OverlayNode) {
			nodeGUID := guid.NewGUIDBytes(overlayNeighbor.GUID())
			nodeResourcesHandled := t.resourcesMap.ResourcesByGUID(*nodeGUID)

			if isValidNeighbor(nodeGUID) && t.handledResources.Equals(*nodeResourcesHandled) {
				t.client.AdvertiseOffersNeighbor( // Sends advertise local offers message
					context.Background(),
					&types.Node{GUID: t.guid.String()},
					&types.Node{IP: overlayNeighbor.IP(), GUID: nodeGUID.String()},
					traderOffering)
			}
		}

		if t.config.Simulation() {
			advertise(overlayNeighbor)
		} else {
			go advertise(overlayNeighbor)
		}
	}
}

// ======================= External Services (Consumed during simulation ONLY) =========================

//Simulation
func (t *Trader) RefreshOffersSim() {
	t.offersMutex.Lock()
	defer t.offersMutex.Unlock()

	for _, offer := range t.offers {
		if offer.Refresh() {
			refreshed, err := t.client.RefreshOffer(
				context.Background(),
				&types.Node{GUID: t.guid.String()},
				&types.Node{IP: offer.supplierIP},
				&types.Offer{ID: int64(offer.ID())},
			)

			offerKEY := offerKey{supplierIP: offer.supplierIP, id: common.OfferID(offer.ID())}
			offer, exist := t.offers[offerKEY]

			if err == nil && refreshed && exist { // Offer exist and was refreshed
				log.Debugf(util.LogTag("TRADER")+"Refresh SUCCEEDED ,supplier: %s, offer: %d",
					offer.SupplierIP(), offer.ID())
				offer.RefreshSucceeded()
			} else if err == nil && !refreshed && exist { // Offer did not exist, so it was not refreshed
				log.Debugf(util.LogTag("TRADER")+"Refresh FAILED (offer did not exist),"+
					" supplier: %s, offer: %d", offer.SupplierIP(), offer.ID())
				delete(t.offers, offerKEY)
			} else if err != nil && exist { // Offer exist but the refresh message failed
				log.Debugf(util.LogTag("TRADER")+"Refresh FAILED, supplier: %s, offer: %d",
					offer.SupplierIP(), offer.ID())
				offer.RefreshFailed()
				if offer.RefreshesFailed() >= t.config.MaxRefreshesFailed() {
					log.Debugf(util.LogTag("TRADER")+"REMOVING offer, supplier: %s, offer: %d",
						offer.SupplierIP(), offer.ID())
					delete(t.offers, offerKEY)
				}
			}
		}
	}
}

//Simulation
func (t *Trader) SpreadOffersSim() {
	// Advertise offers (if any) into the neighbors traders.
	// Necessary only to overcame the problems of unnoticed death of a neighbor.
	if t.haveOffers() && t.config.SpreadOffers() {
		t.advertiseOffersToNeighbors(func(neighborGUID *guid.GUID) bool { return true },
			&types.Node{GUID: t.guid.String(), IP: t.config.HostIP()})
	}
}

// ===============================================================================
// =							SubComponent Interface                           =
// ===============================================================================

func (t *Trader) Start() {
	t.Started(t.config.Simulation(), func() {
		if !t.config.Simulation() {
			go t.start()
		}
	})
}

func (t *Trader) Stop() {
	t.Stopped(func() {
		t.quitChan <- true
	})
}

func (t *Trader) IsWorking() bool {
	return t.Working()
}
