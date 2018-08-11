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
	resourcesMap     *resources.Mapping   // GUID<->Resources mapping
	handledResources *resources.Resources // Combination of resources that its responsible for managing (FIXED)

	nearbyTradersOffering *NearbyTradersOffering    // Nearby traders that might have offers available
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

		nearbyTradersOffering: NewNeighborTradersOffering(),
		offers:                make(map[offerKey]*traderOffer),
		offersMutex:           &sync.Mutex{},

		quitChan:            make(chan bool),
		refreshOffersTicker: time.NewTicker(config.RefreshingInterval()).C,
		spreadOffersTimer:   time.NewTicker(config.SpreadOffersInterval()).C,
	}
}

// Runs a endless loop goroutine that dispatches timer events into other goroutines.
func (trader *Trader) start() {
	for {
		select {
		case <-trader.refreshOffersTicker: // Time to refresh all the current offers (verify if the suppliers are alive)
			trader.offersMutex.Lock()

			for _, offer := range trader.offers {
				if offer.Refresh() {
					go func(offer *traderOffer) {
						refreshed, err := trader.client.RefreshOffer(
							context.Background(),
							&types.Node{GUID: trader.guid.String()},
							&types.Node{IP: offer.supplierIP},
							&types.Offer{ID: int64(offer.ID())},
						)

						trader.offersMutex.Lock()
						defer trader.offersMutex.Unlock()

						offerKEY := offerKey{supplierIP: offer.supplierIP, id: common.OfferID(offer.ID())}
						offer, exist := trader.offers[offerKEY]

						if err == nil && refreshed && exist { // Offer exist and was refreshed
							log.Debugf(util.LogTag("TRADER")+"Refresh SUCCEEDED ,supplier: %s, offer: %d",
								offer.SupplierIP(), offer.ID())
							offer.RefreshSucceeded()
						} else if err == nil && !refreshed && exist { // Offer did not exist, so it was not refreshed
							log.Debugf(util.LogTag("TRADER")+"Refresh FAILED (offer did not exist),"+
								" supplier: %s, offer: %d", offer.SupplierIP(), offer.ID())
							delete(trader.offers, offerKEY)
						} else if err != nil && exist { // Offer exist but the refresh message failed
							log.Debugf(util.LogTag("TRADER")+"Refresh FAILED, supplier: %s, offer: %d",
								offer.SupplierIP(), offer.ID())
							offer.RefreshFailed()
							if offer.RefreshesFailed() >= trader.config.MaxRefreshesFailed() {
								log.Debugf(util.LogTag("TRADER")+"REMOVING offer, supplier: %s, offer: %d",
									offer.SupplierIP(), offer.ID())
								delete(trader.offers, offerKEY)
							}
						}
					}(offer)
				}
			}

			trader.offersMutex.Unlock()
		case <-trader.spreadOffersTimer:
			// Advertise offers (if any) into the neighbors traders.
			// Necessary only to overcame the problems of unnoticed death of a neighbor.
			if trader.haveOffers() {
				trader.advertiseOffersToNeighbors(func(neighborGUID *guid.GUID) bool { return true })
			}
		case quit := <-trader.quitChan: // Stopping the trader (returning the goroutine)
			if quit {
				log.Infof(util.LogTag("TRADER")+"Trader %s STOPPED", trader.guid.Short())
				return
			}
		}
	}
}

// Returns all the offers that the trader is managing.
func (trader *Trader) GetOffers(ctx context.Context, fromNode *types.Node, relay bool) []types.AvailableOffer {
	if trader.haveOffers() { // Trader has offers so return them immediately
		trader.offersMutex.Lock()
		defer trader.offersMutex.Unlock()

		availableOffers := len(trader.offers)
		allOffers := make([]types.AvailableOffer, availableOffers)
		index := 0
		for _, traderOffer := range trader.offers {
			allOffers[index].SupplierIP = traderOffer.SupplierIP()
			allOffers[index].ID = int64(traderOffer.ID())
			allOffers[index].Amount = traderOffer.Amount()
			allOffers[index].Resources = types.Resources{
				CPUs: traderOffer.Resources().CPUs(),
				RAM:  traderOffer.Resources().RAM(),
			}
			index++
		}
		return allOffers
	} else { // Ask for offers in the nearby neighbors that we think they have offers (via offer advertise relaying)
		resOffers := make([]types.AvailableOffer, 0)
		fromNodeGuid := guid.NewGUIDString(fromNode.GUID)

		// Ask the successor (higher GUID)
		successor := trader.nearbyTradersOffering.Successor()
		if successor != nil && (relay || (!relay && (trader.guid.Higher(*fromNodeGuid)))) {
			successorResourcesHandled := trader.resourcesMap.ResourcesByGUID(*successor.GUID())
			if trader.handledResources.Equals(*successorResourcesHandled) {
				offers, err := trader.client.GetOffers(
					ctx,
					&types.Node{
						GUID: trader.guid.String(),
					},
					&types.Node{
						IP:   successor.IP(),
						GUID: successor.GUID().String(),
					},
					false,
				) // Sends the get offers message
				if err == nil && offers != nil {
					resOffers = append(resOffers, offers...)
				}
			}

		}

		// Ask the predecessor (lower GUID)
		predecessor := trader.nearbyTradersOffering.Predecessor()
		if predecessor != nil && (relay || (!relay && (trader.guid.Lower(*fromNodeGuid)))) {
			predecessorResourcesHandled := trader.resourcesMap.ResourcesByGUID(*predecessor.GUID())
			if trader.handledResources.Equals(*predecessorResourcesHandled) {
				offers, err := trader.client.GetOffers(
					ctx,
					&types.Node{GUID: trader.guid.String()},
					&types.Node{IP: predecessor.IP(), GUID: predecessor.GUID().String()},
					false,
				) // Sends the get offers message
				if err == nil && offers != nil {
					resOffers = append(resOffers, offers...)
				}
			}

		}
		// TRY: OPTIONAl make the calls in parallel (2 goroutines) and wait here for both, then join the results.
		return resOffers
	}
}

// Receives a resource offer from other node (supplier) of the system
func (trader *Trader) CreateOffer(fromSupp *types.Node, newOffer *types.Offer) {
	resourcesOffered := resources.NewResources(newOffer.Resources.CPUs, newOffer.Resources.RAM)

	// Verify if the offer contains the resources of trader.
	// Basically verify if the offer is bigger than the handled resources of the trader.
	if resourcesOffered.Contains(*trader.handledResources) {
		trader.offersMutex.Lock()

		offerKey := offerKey{supplierIP: fromSupp.IP, id: common.OfferID(newOffer.ID)}
		offer := newTraderOffer(*guid.NewGUIDString(fromSupp.GUID), fromSupp.IP, common.OfferID(newOffer.ID),
			newOffer.Amount, *resourcesOffered)

		advertise := len(trader.offers) == 0

		trader.offers[offerKey] = offer
		log.Debugf(util.LogTag("TRADER")+"%s Offer CREATED %dX<%d;%d>, From: %s, Offer: %d",
			trader.guid.Short(), newOffer.Amount, newOffer.Resources.CPUs, newOffer.Resources.RAM,
			fromSupp.IP, newOffer.ID)

		trader.offersMutex.Unlock()

		if advertise { // If node had no offers, advertise it has now for all the neighbors
			if trader.config.Simulation() {
				trader.advertiseOffersToNeighbors(func(neighborGUID *guid.GUID) bool { return true })
			} else {
				go trader.advertiseOffersToNeighbors(func(neighborGUID *guid.GUID) bool { return true })
			}
		}
	}
}

// Remove an offer from the offering table.
func (trader *Trader) RemoveOffer(fromSupp *types.Node, offer *types.Offer) {
	trader.offersMutex.Lock()
	defer trader.offersMutex.Unlock()

	delete(trader.offers, offerKey{supplierIP: fromSupp.IP, id: common.OfferID(offer.ID)})

	log.Debugf(util.LogTag("TRADER")+"Offer REMOVED Offer: %d, Supp: %s", offer.ID, fromSupp.IP)
}

// Relay the offering advertise for the overlay neighbors if the trader doesn't have any available offers
func (trader *Trader) AdvertiseNeighborOffer(fromTrader, toNeighborTrader, traderOffering *types.Node) {
	fromTraderGuid := guid.NewGUIDString(fromTrader.GUID)
	traderOfferingGuid := guid.NewGUIDString(traderOffering.GUID)

	var isValidNeighbor func(neighborGUID *guid.GUID) bool = nil
	if trader.guid.Cmp(*fromTraderGuid) > 0 { // Message comes from a lower GUID's node
		trader.nearbyTradersOffering.SetPredecessor(nodeCommon.NewRemoteNode(traderOffering.IP, *traderOfferingGuid))
		// Relay the advertise to a higher GUID's node (inside partition)
		isValidNeighbor = func(neighborGUID *guid.GUID) bool {
			return neighborGUID.Higher(*trader.guid)
		}
	} else { // Message comes from a higher GUID's node
		trader.nearbyTradersOffering.SetPredecessor(nodeCommon.NewRemoteNode(traderOffering.IP, *traderOfferingGuid))
		// Relay the advertise to a lower GUID's node (inside partition)
		isValidNeighbor = func(neighborGUID *guid.GUID) bool {
			return neighborGUID.Lower(*trader.guid)
		}
	}

	// Do not relay the advertise if the node has offers.
	if !trader.haveOffers() {
		if trader.config.Simulation() {
			trader.advertiseOffersToNeighbors(isValidNeighbor)
		} else {
			go trader.advertiseOffersToNeighbors(isValidNeighbor)
		}
	}
}

func (trader *Trader) haveOffers() bool {
	trader.offersMutex.Lock()
	defer trader.offersMutex.Unlock()

	return len(trader.offers) != 0
}

// Advertise that we have offers into all trader's neighbors that survive the given predicate application.
func (trader *Trader) advertiseOffersToNeighbors(isValidNeighbor func(neighborGUID *guid.GUID) bool) {
	log.Debugf(util.LogTag("TRADER")+"ADVERTISE offers, From: %s", trader.guid.Short())

	overlayNeighbors, err := trader.overlay.Neighbors(context.Background(), trader.guid.Bytes())
	if err != nil {
		return
	}

	for _, overlayNeighbor := range overlayNeighbors { // Advertise to all neighbors (inside resource partition)
		advertise := func(overlayNeighbor *overlayTypes.OverlayNode) {
			nodeGUID := guid.NewGUIDBytes(overlayNeighbor.GUID())
			nodeResourcesHandled := trader.resourcesMap.ResourcesByGUID(*nodeGUID)

			if isValidNeighbor(nodeGUID) && trader.handledResources.Equals(*nodeResourcesHandled) {
				trader.client.AdvertiseOffersNeighbor(
					context.Background(),
					&types.Node{GUID: trader.guid.String()},
					&types.Node{IP: overlayNeighbor.IP(), GUID: nodeGUID.String()},
					&types.Node{IP: trader.config.Host.IP, GUID: trader.guid.String()},
				) // Sends advertise local offers message
			}
		}

		if trader.config.Simulation() {
			advertise(overlayNeighbor)
		} else {
			go advertise(overlayNeighbor)
		}
	}
}

// ======================= External Services (Consumed during simulation ONLY) =========================

//Simulation
func (trader *Trader) RefreshOffersSim() {
	trader.offersMutex.Lock()
	defer trader.offersMutex.Unlock()

	for _, offer := range trader.offers {
		if offer.Refresh() {
			refreshed, err := trader.client.RefreshOffer(
				context.Background(),
				&types.Node{GUID: trader.guid.String()},
				&types.Node{IP: offer.supplierIP},
				&types.Offer{ID: int64(offer.ID())},
			)

			offerKEY := offerKey{supplierIP: offer.supplierIP, id: common.OfferID(offer.ID())}
			offer, exist := trader.offers[offerKEY]

			if err == nil && refreshed && exist { // Offer exist and was refreshed
				log.Debugf(util.LogTag("TRADER")+"Refresh SUCCEEDED ,supplier: %s, offer: %d",
					offer.SupplierIP(), offer.ID())
				offer.RefreshSucceeded()
			} else if err == nil && !refreshed && exist { // Offer did not exist, so it was not refreshed
				log.Debugf(util.LogTag("TRADER")+"Refresh FAILED (offer did not exist),"+
					" supplier: %s, offer: %d", offer.SupplierIP(), offer.ID())
				delete(trader.offers, offerKEY)
			} else if err != nil && exist { // Offer exist but the refresh message failed
				log.Debugf(util.LogTag("TRADER")+"Refresh FAILED, supplier: %s, offer: %d",
					offer.SupplierIP(), offer.ID())
				offer.RefreshFailed()
				if offer.RefreshesFailed() >= trader.config.MaxRefreshesFailed() {
					log.Debugf(util.LogTag("TRADER")+"REMOVING offer, supplier: %s, offer: %d",
						offer.SupplierIP(), offer.ID())
					delete(trader.offers, offerKEY)
				}
			}
		}
	}
}

//Simulation
func (trader *Trader) SpreadOffersSim() {
	// Advertise offers (if any) into the neighbors traders.
	// Necessary only to overcame the problems of unnoticed death of a neighbor.
	if trader.haveOffers() {
		trader.advertiseOffersToNeighbors(func(neighborGUID *guid.GUID) bool { return true })
	}
}

// ===============================================================================
// =							SubComponent Interface                           =
// ===============================================================================

func (trader *Trader) Start() {
	trader.Started(trader.config.Simulation(), func() {
		if !trader.config.Simulation() {
			go trader.start()
		}
	})
}

func (trader *Trader) Stop() {
	trader.Stopped(func() {
		trader.quitChan <- true
	})
}

func (trader *Trader) IsWorking() bool {
	return trader.Working()
}
