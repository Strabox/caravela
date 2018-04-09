package trader

import (
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/api/remote"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/api"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/discovery/common"
	"sync"
	"time"
)

/*
Trader is responsible for managing offers from suppliers and negotiate these offers with buyers
*/
type Trader struct {
	config           *configuration.Configuration // System configuration values
	client           remote.Caravela              // Client for the system
	guid             *guid.Guid                   // Trader's own GUID
	handledResources *resources.Resources         // Combination of resources that its responsible for managing (Static value)

	refreshOffersTicker <-chan time.Time // Time ticker for running the refreshing offer messages

	offers      map[offerKey]*traderOffer // Map with all the offers that the trader is managing
	offersMutex *sync.Mutex               // Mutex for managing the offer
}

func NewTrader(config *configuration.Configuration, client remote.Caravela, guid guid.Guid,
	resources resources.Resources) *Trader {
	res := &Trader{}
	res.config = config
	res.client = client
	res.guid = &guid
	res.handledResources = &resources

	res.refreshOffersTicker = time.NewTicker(config.RefreshingInterval()).C

	res.offers = make(map[offerKey]*traderOffer)
	res.offersMutex = &sync.Mutex{}
	return res
}

func (trader *Trader) Guid() *guid.Guid {
	return trader.guid.Copy()
}

func (trader *Trader) Start() {
	log.Debugln("[Trader] Starting refreshing resource's offers...")
	go trader.refreshingOffers()
}

/*
Runs in a individual goroutine and refreshes the trader's offers from time to time
*/
func (trader *Trader) refreshingOffers() {
	for {
		select {
		case <-trader.refreshOffersTicker: // Time to refresh all the current offers (verify if the suppliers are alive)
			trader.offersMutex.Lock()

			for _, offer := range trader.offers {
				if offer.Refresh() {
					go func() {
						trader.offersMutex.Lock()
						defer trader.offersMutex.Unlock()

						_, refreshed := trader.client.RefreshOffer(offer.supplierIP, trader.guid.String(), int64(offer.ID()))

						offerKEY := offerKey{common.OfferID(offer.ID()), offer.supplierIP}
						offer, exist := trader.offers[offerKEY]

						if refreshed && exist {
							log.Debugf("[Trader] Refresh SUCCEEDED for supplier: %s offer: %d", offer.SupplierIP(), offer.ID())
							offer.RefreshSucceeded()
						} else if !refreshed && exist {
							log.Debugf("[Trader] Refresh FAILED for supplier: %s offer: %d", offer.SupplierIP(), offer.ID())
							offer.RefreshFailed()
							if offer.refreshesFailed >= trader.config.MaxRefreshesFailed() {
								log.Debugf("[Trader] Removing offer of supplier: %s offer: %d", offer.SupplierIP(), offer.ID())
								delete(trader.offers, offerKEY)
							}
						}
					}()
				}
			}

			trader.offersMutex.Unlock()
		}
	}
}

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
	log.Debugf("[Trader] %s offer received %dX(CPUs: %d, RAM: %d) from: %s", trader.guid.String(), amount, cpus, ram, supplierIP)

	resourcesOffered := resources.NewResources(cpus, ram)
	// Verify if the offer contains the resources of trader (Basically if the offer is bigger than the handled resources)
	if resourcesOffered.Contains(*trader.handledResources) {
		trader.offersMutex.Lock()
		defer trader.offersMutex.Unlock()

		offer := newTraderOffer(*guid.NewGuidString(supplierGUID), supplierIP, common.NewOffer(common.OfferID(id),
			amount, *resources.NewResources(cpus, ram)))
		offerKey := offerKey{common.OfferID(id), supplierIP}

		trader.offers[offerKey] = offer
		log.Debugf("[Trader] %s Offer CREATED %dX(CPUs: %d, RAM: %d) from: %s", trader.guid.String(), amount, cpus, ram, supplierIP)
	}
}

func (trader *Trader) RemoveOffer(fromSupplierIP string, fromSupplierGUID string, toTraderGUID string, offerID int64) {
	// TODO
}
