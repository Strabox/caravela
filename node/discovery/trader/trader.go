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

	guid             *guid.GUID                // Trader's own GUID
	resourcesMap     *resources.Mapping        // GUID<->Resources mapping
	handledResources *resources.Resources      // Combination of resources that its responsible for managing (FIXED)
	offers           map[offerKey]*traderOffer // Map with all the offers that the trader is managing
	offersMutex      *sync.Mutex               // Mutex for managing the offers

	quitChan            chan bool        // Channel to alert that the node is stopping
	refreshOffersTicker <-chan time.Time // Time ticker for sending the refreshing offer messages
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

	res.quitChan = make(chan bool)
	res.refreshOffersTicker = time.NewTicker(config.RefreshingInterval()).C

	res.offers = make(map[offerKey]*traderOffer)
	res.offersMutex = &sync.Mutex{}
	return res
}

/*
Runs in an individual goroutine and refreshes the trader's offers from time to time.
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
						go func() {
							trader.offersMutex.Lock()
							defer trader.offersMutex.Unlock()

							err, refreshed := trader.client.RefreshOffer(offer.supplierIP, trader.guid.String(),
								int64(offer.ID()))

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
						}()
					}
				}
			}()
		case res := <-trader.quitChan: // Stopping the trader
			if res {
				log.Infof(util.LogTag("Trader")+"Trader %s STOPPED", trader.guid.String())
				return
			}
		}
	}
}

/*
Returns all the offers that the trader is managing.
*/
func (trader *Trader) GetOffers() []api.Offer {
	trader.offersMutex.Lock()
	defer trader.offersMutex.Unlock()

	availableOffers := len(trader.offers)
	if len(trader.offers) <= 0 {
		return nil
	} else {
		resOffers := make([]api.Offer, availableOffers)
		index := 0
		for _, offer := range trader.offers {
			resOffers[index].SupplierIP = trader.config.HostIP()
			resOffers[index].ID = int64(offer.ID())
			index++
		}
		return resOffers
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
		log.Debugf(util.LogTag("Trader")+"%s Offer CREATED %dX(CPUs: %d, RAM: %d) from: %s",
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
	log.Debugf(util.LogTag("Trader")+"Removing offer of supplier: %s offer: %d", fromSupplierIP, offerID)
}

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

func (trader *Trader) Guid() *guid.GUID {
	return trader.guid.Copy()
}

func (trader *Trader) HandledResources() *resources.Resources {
	return trader.handledResources.Copy()
}
