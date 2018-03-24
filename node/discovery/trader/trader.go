package trader

import (
	"github.com/strabox/caravela/api/client"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/discovery/common"
	"github.com/strabox/caravela/node/guid"
	"github.com/strabox/caravela/node/resources"
	"log"
	"sync"
	"time"
)

/*
offerKey is based on the local offer id and the supplier IP (Hoping the supplier IP is unique????)
*/
type offerKey struct {
	id         common.OfferID
	supplierIP string
}

/*
Trader is responsible for managing offers from suppliers and negotiate these offers with buyers
*/
type Trader struct {
	config           *configuration.Configuration // System configuration values
	client           client.Caravela              // Client for the system
	guid             *guid.Guid                   // Trader's own GUID
	handledResources *resources.Resources         // Combination of resources that its responsible for managing

	refreshOffersTicker <-chan time.Time // Time ticker for running the refreshing offer messages

	offersMap   map[offerKey]*traderOffer // Map with all the offers that the trader is managing
	offersMutex *sync.Mutex               // Mutex for managing the offer
}

func NewTrader(config *configuration.Configuration, client client.Caravela, guid guid.Guid,
	resources resources.Resources) *Trader {
	res := &Trader{}
	res.config = config
	res.client = client
	res.guid = &guid
	res.handledResources = &resources

	res.refreshOffersTicker = time.NewTicker(config.RefreshingInterval()).C

	res.offersMap = make(map[offerKey]*traderOffer)
	res.offersMutex = &sync.Mutex{}
	return res
}

func (trader *Trader) Guid() *guid.Guid {
	return trader.guid.Copy()
}

func (trader *Trader) Start() {
	log.Println("[Trader] Starting refreshing resource offers...")
	go trader.refreshingOffers()
}

/*
Runs in a individual thread and refreshes the trader's offers from time to time
*/
func (trader *Trader) refreshingOffers() {
	responsesChan := make(chan client.OfferRefreshResponse)
	for {
		select {
		case <-trader.refreshOffersTicker:
			trader.offersMutex.Lock()

			for _, offer := range trader.offersMap {
				if offer.Refresh() {
					go trader.client.RefreshOffer(offer.supplierIP, trader.guid.String(), int(offer.LocalID()), responsesChan)
				}
			}

			trader.offersMutex.Unlock()
		case response := <-responsesChan:
			trader.offersMutex.Lock()

			offerKEY := offerKey{common.OfferID(response.OfferID), response.ToSupplierIP}
			offer, exist := trader.offersMap[offerKEY]

			if response.Success && exist {
				log.Printf("[Trader] Refresh SUCCEEDED for supplier: %s offer: %d\n", offer.SupplierIP(), offer.LocalID())
				offer.RefreshSucceeded()
			} else if !response.Success && exist {
				log.Printf("[Trader] Refresh FAILED for supplier: %s offer: %d\n", offer.SupplierIP(), offer.LocalID())
				offer.RefreshFailed()
				if offer.refreshesFailed >= trader.config.MaxRefreshesFailed() {
					log.Printf("[Trader] Removing offer of supplier: %s offer: %d\n", offer.SupplierIP(), offer.LocalID())
					delete(trader.offersMap, offerKEY)
				}
			}
			trader.offersMutex.Unlock()
		}
	}
}

/*
Receives a resource offer from other node of the system
*/
func (trader *Trader) CreateOffer(id int64, amount int, cpus int, ram int, supplierGUID string, supplierIP string) {
	log.Printf("[Trader] %s offer received %dX(CPUs: %d, RAM: %d) from: %s\n", trader.guid.String(), amount, cpus, ram, supplierIP)

	// Verify if this trader is responsible for these type of offers
	//if cpus == trader.handledResources.CPU() && ram == trader.handledResources.RAM() {
	trader.offersMutex.Lock()
	defer trader.offersMutex.Unlock()

	offer := newTraderOffer(*guid.NewGuidString(supplierGUID), supplierIP, *common.NewOffer(common.OfferID(id),
		amount, *resources.NewResources(cpus, ram)))
	offerKey := offerKey{common.OfferID(id), supplierIP}

	trader.offersMap[offerKey] = offer
	//}
}
