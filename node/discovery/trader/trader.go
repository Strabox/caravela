package trader

import (
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/api/remote"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/api"
	nodeCommon "github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/discovery/common"
	"github.com/strabox/caravela/overlay"
	"github.com/strabox/caravela/util"
	"sync"
	"time"
)

/*
Trader is responsible for managing offers from multiple suppliers and negotiate these offers with buyers.
The resources combination that the trader can handle is described by its GUID.
*/
type Trader struct {
	nodeCommon.SystemSubComponent // Base component

	config  *configuration.Configuration // System's configurations
	overlay overlay.Overlay              // Node overlay to efficient route messages to specific nodes.
	client  remote.Caravela              // Client for the system

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

func NewTrader(config *configuration.Configuration, overlay overlay.Overlay, client remote.Caravela,
	guid guid.GUID, resourcesMapping *resources.Mapping) *Trader {
	res := &Trader{}
	res.config = config
	res.overlay = overlay
	res.client = client
	res.guid = &guid
	res.resourcesMap = resourcesMapping
	res.handledResources, _ = res.resourcesMap.ResourcesByGUID(*res.guid)

	res.nearbyTradersOffering = NewNeighborTradersOffering()
	res.offers = make(map[offerKey]*traderOffer)
	res.offersMutex = &sync.Mutex{}

	res.quitChan = make(chan bool)
	res.refreshOffersTicker = time.NewTicker(config.RefreshingInterval()).C
	res.spreadOffersTimer = time.NewTicker(config.SpreadOffersInterval()).C
	return res
}

/*
Runs a endless loop goroutine that dispatches timer events into other goroutines.
*/
func (trader *Trader) refreshingOffers() {
	for {
		select {
		case <-trader.refreshOffersTicker: // Time to refresh all the current offers (verify if the suppliers are alive)
			go func() {
				trader.offersMutex.Lock()
				defer trader.offersMutex.Unlock()

				for _, offer := range trader.offers {
					if offer.Refresh() {
						go func(offer *traderOffer) {
							refreshed, err := trader.client.RefreshOffer(offer.supplierIP, trader.guid.String(),
								int64(offer.ID())) // Sends refresh message

							trader.offersMutex.Lock()
							defer trader.offersMutex.Unlock()

							offerKEY := offerKey{supplierIP: offer.supplierIP, id: common.OfferID(offer.ID())}
							offer, exist := trader.offers[offerKEY]

							if err == nil && refreshed && exist { // Offer exist and was refreshed
								log.Debugf(util.LogTag("Trader")+"Refresh SUCCEEDED for supplier: %s offer: %d",
									offer.SupplierIP(), offer.ID())
								offer.RefreshSucceeded()
							} else if err == nil && !refreshed && exist { // Offer did not exist, so it was not refreshed
								log.Debugf(util.LogTag("Trader")+"Refresh FAILED (offer did not exist)"+
									" for supplier: %s offer: %d", offer.SupplierIP(), offer.ID())
								delete(trader.offers, offerKEY)
							} else if err != nil && exist { // Offer exist but the refresh message failed
								log.Debugf(util.LogTag("Trader")+"Refresh FAILED for supplier: %s offer: %d",
									offer.SupplierIP(), offer.ID())
								offer.RefreshFailed()
								if offer.RefreshesFailed() >= trader.config.MaxRefreshesFailed() {
									log.Debugf(util.LogTag("Trader")+"Removing offer of supplier: %s offer: %d",
										offer.SupplierIP(), offer.ID())
									delete(trader.offers, offerKEY)
								}
							}
						}(offer)
					}
				}
			}()
		case <-trader.spreadOffersTimer: // Advertise offers (if any) into the neighbors traders
			go func() {
				if !trader.haveOffers() {
					return
				} // If trader has no offer don't advertise to neighbors

				// TODO: Verify if necessary cause this makes a lookup happen in Chord?
				overlayNeighbors, err := trader.overlay.Neighbors(trader.guid.Bytes())
				if err != nil {
					return
				}

				for _, overlayNeighbor := range overlayNeighbors { // Advertise to the lower and higher GUID's node (inside partition)
					go func(overlayNeighbor *overlay.Node) {
						nodeGUID := guid.NewGUIDBytes(overlayNeighbor.GUID())
						nodeResourcesHandled, _ := trader.resourcesMap.ResourcesByGUID(*nodeGUID)
						if trader.handledResources.Equals(*nodeResourcesHandled) {
							trader.client.AdvertiseOffersNeighbor(overlayNeighbor.IP(), nodeGUID.String(),
								trader.guid.String(), trader.guid.String(), trader.config.Host.IP) // Sends advertise local offers message
						}
					}(overlayNeighbor)
				}
			}()
		case quit := <-trader.quitChan: // Stopping the trader (returning the goroutine)
			if quit {
				log.Infof(util.LogTag("Trader")+"Trader %s STOPPED", trader.guid.String())
				return
			}
		}
	}
}

/*
Returns all the offers that the trader is managing.
*/
func (trader *Trader) GetOffers(relay bool, fromNodeGUID string) []api.Offer {
	if trader.haveOffers() { // Trader has offers so return them immediately
		trader.offersMutex.Lock()
		defer trader.offersMutex.Unlock()

		availableOffers := len(trader.offers)
		resOffers := make([]api.Offer, availableOffers)
		index := 0
		for _, offer := range trader.offers {
			resOffers[index].SupplierIP = trader.config.HostIP()
			resOffers[index].ID = int64(offer.ID())
			index++
		}
		return resOffers
	} else { // Ask for offers in the nearby neighbors that we think they have offers (via offer advertise relaying)
		res := make([]api.Offer, 0)
		fromNodeGuid := guid.NewGUIDString(fromNodeGUID)

		// Ask the successor (higher GUID)
		successor := trader.nearbyTradersOffering.Successor()
		if successor != nil && (relay || (!relay && (trader.guid.Cmp(*fromNodeGuid) > 0))) {
			successorResourcesHandled, _ := trader.resourcesMap.ResourcesByGUID(*successor.GUID())
			if trader.handledResources.Equals(*successorResourcesHandled) {
				offers, err := trader.client.GetOffers(successor.IP(), successor.GUID().String(),
					false, trader.guid.String()) // Sends the get offers message
				if err == nil && offers != nil {
					res = append(res, offers...)
				}
			}

		}

		// Ask the predecessor (lower GUID)
		predecessor := trader.nearbyTradersOffering.Predecessor()
		if predecessor != nil && (relay || (!relay && (trader.guid.Cmp(*fromNodeGuid) < 0))) {
			predecessorResourcesHandled, _ := trader.resourcesMap.ResourcesByGUID(*predecessor.GUID())
			if trader.handledResources.Equals(*predecessorResourcesHandled) {
				offers, err := trader.client.GetOffers(predecessor.IP(), predecessor.GUID().String(),
					false, trader.guid.String()) // Sends the get offers message
				if err == nil && offers != nil {
					res = append(res, offers...)
				}
			}

		}

		// TODO: OPTIONAl make the calls in parallel (2 goroutines) and wait here for both, then join the results.
		return res
	}
}

/*
Receives a resource offer from other node (supplier) of the system
*/
func (trader *Trader) CreateOffer(id int64, amount int, cpus int, ram int, supplierGUID string, supplierIP string) {
	resourcesOffered := resources.NewResources(cpus, ram)

	// Verify if the offer contains the resources of trader.
	// Basically verify if the offer is bigger than the handled resources of the trader.
	if resourcesOffered.Contains(*trader.handledResources) {
		trader.offersMutex.Lock()
		defer trader.offersMutex.Unlock()

		offer := newTraderOffer(*guid.NewGUIDString(supplierGUID), supplierIP, common.OfferID(id),
			amount, *resources.NewResources(cpus, ram))
		offerKey := offerKey{supplierIP: supplierIP, id: common.OfferID(id)}

		trader.offers[offerKey] = offer
		log.Debugf(util.LogTag("Trader")+"%s OFFER CREATED %dX<%d;%d>, From: %s",
			trader.guid.String(), amount, cpus, ram, supplierIP)
	}
}

/*
Remove an offer from the "advertising" table.
*/
func (trader *Trader) RemoveOffer(fromSupplierIP string, fromSupplierGUID string, toTraderGUID string, offerID int64) {
	trader.offersMutex.Lock()
	defer trader.offersMutex.Unlock()

	delete(trader.offers, offerKey{supplierIP: fromSupplierIP, id: common.OfferID(offerID)})
	log.Debugf(util.LogTag("Trader")+"OFFER REMOVED %d, Supp: %s", offerID, fromSupplierIP)
}

/*
Relay the offering advertise for the overlay neighbors if the trader doesn't have any available offers
*/
func (trader *Trader) AdvertiseNeighborOffer(fromTraderGUID string, traderOfferingIP string, traderOfferingGUID string) {
	if trader.haveOffers() {
		return
	}
	fromTraderGuid := guid.NewGUIDString(fromTraderGUID)
	traderOfferingGuid := guid.NewGUIDString(traderOfferingGUID)

	// TODO: OPTIONAL try refactor this two ifs (due to code duplication)
	if trader.guid.Cmp(*fromTraderGuid) > 0 { // Message comes from a lower GUID's node
		trader.nearbyTradersOffering.SetPredecessor(nodeCommon.NewRemoteNode(traderOfferingIP, *traderOfferingGuid))

		// TODO: Verify if necessary cause this makes a lookup happen in Chord?
		overlayNeighbors, err := trader.overlay.Neighbors(trader.guid.Bytes())
		if err != nil {
			return
		}

		for _, overlayNeighbor := range overlayNeighbors {
			go func(overlayNeighbor *overlay.Node) {
				nodeGUID := guid.NewGUIDBytes(overlayNeighbor.GUID())
				nodeResourcesHandled, _ := trader.resourcesMap.ResourcesByGUID(*nodeGUID)
				//log.Debugf(util.LogTag("Trader")+"AdvertiseNeighbor: %s", nodeGUID.String())
				if nodeGUID.Cmp(*trader.guid) > 0 && trader.handledResources.Equals(*nodeResourcesHandled) {
					// Relay the advertise to a higher GUID's node (inside partition)
					trader.client.AdvertiseOffersNeighbor(overlayNeighbor.IP(), nodeGUID.String(), trader.guid.String(),
						traderOfferingIP, traderOfferingGUID)
				}
			}(overlayNeighbor)
		}
	} else { // Message comes from a higher GUID's node
		trader.nearbyTradersOffering.SetSuccessor(nodeCommon.NewRemoteNode(traderOfferingIP, *traderOfferingGuid))

		// TODO: Verify if necessary cause this makes a lookup happen in Chord?
		overlayNeighbors, err := trader.overlay.Neighbors(trader.guid.Bytes())
		if err != nil {
			return
		}

		for _, overlayNeighbor := range overlayNeighbors {
			go func(overlayNeighbor *overlay.Node) {
				nodeGUID := guid.NewGUIDBytes(overlayNeighbor.GUID())
				nodeResourcesHandled, _ := trader.resourcesMap.ResourcesByGUID(*nodeGUID)
				//log.Debugf(util.LogTag("Trader")+"AdvertiseNeighbor: %s", nodeGUID.String())
				if nodeGUID.Cmp(*trader.guid) < 0 && trader.handledResources.Equals(*nodeResourcesHandled) {
					// Relay the advertise to a lower GUID's node (inside partition)
					trader.client.AdvertiseOffersNeighbor(overlayNeighbor.IP(), nodeGUID.String(), trader.guid.String(),
						traderOfferingIP, traderOfferingGUID)
				}
			}(overlayNeighbor)
		}
	}
}

func (trader *Trader) GUID() *guid.GUID {
	return trader.guid.Copy()
}

func (trader *Trader) HandledResources() *resources.Resources {
	return trader.handledResources.Copy()
}

func (trader *Trader) haveOffers() bool {
	trader.offersMutex.Lock()
	defer trader.offersMutex.Unlock()

	return len(trader.offers) != 0
}

/*
===============================================================================
							SubComponent Interface
===============================================================================
*/

func (trader *Trader) Start() {
	trader.Started(func() {
		go trader.refreshingOffers()
	})
}

func (trader *Trader) Stop() {
	trader.Stopped(func() {
		trader.quitChan <- true
	})
}

func (trader *Trader) isWorking() bool {
	return trader.Working()
}
