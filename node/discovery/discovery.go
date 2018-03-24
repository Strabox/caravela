package discovery

import (
	"github.com/strabox/caravela/api/client"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/discovery/supplier"
	"github.com/strabox/caravela/node/discovery/trader"
	"github.com/strabox/caravela/node/guid"
	"github.com/strabox/caravela/node/resources"
	"github.com/strabox/caravela/overlay"
	"log"
)

type Discovery struct {
	config       *configuration.Configuration
	client       client.Caravela
	resourcesMap *resources.Mapping

	supplier *supplier.Supplier        // Node supplier offer the node's resources
	traders  map[string]*trader.Trader // Trader help matchmaking offers and searches
}

func NewDiscovery(config *configuration.Configuration, overlay overlay.Overlay,
	client client.Caravela, resourcesMap *resources.Mapping, maxResources resources.Resources) *Discovery {
	res := &Discovery{}
	res.config = config
	res.client = client
	res.resourcesMap = resourcesMap
	res.supplier = supplier.NewSupplier(config, overlay, client, resourcesMap, maxResources)
	res.traders = make(map[string]*trader.Trader)
	return res
}

/*============================== DiscoveryInternal Interface =============================== */

func (disc *Discovery) Start() {
	// Supplier starts supplying resources
	disc.supplier.Start()
}

func (disc *Discovery) AddTrader(traderGUID guid.Guid) {
	traderResources, _ := disc.resourcesMap.ResourcesByGuid(traderGUID)
	newTrader := trader.NewTrader(disc.config, disc.client, traderGUID, *traderResources)
	disc.traders[traderGUID.String()] = newTrader
	newTrader.Start() // Trader starts refreshing offers
	log.Printf("[Discovery] New Trader: %s | Resources: %s\n", (&traderGUID).String(), traderResources.String())
}

func (disc *Discovery) Find() {
	// TODO
}

func (disc *Discovery) Deploy() {
	// TODO
}

/*============================== DiscoveryExternal Interface ============================== */

func (disc *Discovery) CreateOffer(fromSupplierGUID string, fromSupplierIP string, toTraderGUID string, id int,
	amount int, cpus int, ram int) {

	t, exist := disc.traders[toTraderGUID]
	if exist {
		t.CreateOffer(int64(id), amount, cpus, ram, fromSupplierGUID, fromSupplierIP)
	}
}

func (disc *Discovery) RefreshOffer(id int, fromTraderGUID string) bool {
	return disc.supplier.RefreshOffer(id, fromTraderGUID)
}
