package discovery

import (
	"github.com/strabox/caravela/api/client"
	"github.com/strabox/caravela/node/configuration"
	"github.com/strabox/caravela/node/guid"
	"github.com/strabox/caravela/node/resources"
	"github.com/strabox/caravela/overlay"
	"log"
)

type Discovery struct {
	config       *configuration.Configuration
	client       client.Caravela
	resourcesMap *resources.Mapping
	supplier     *supplier
	traders      []*trader
}

/*
Minimum number of traders expected in one CARAVELA's node
*/
const minimumNumberOfTraders = 1

func NewDiscovery(config *configuration.Configuration, overlay overlay.Overlay,
	client client.Caravela, resourcesMap *resources.Mapping,
	maxResources resources.Resources) *Discovery {

	res := &Discovery{}
	res.config = config
	res.client = client
	res.resourcesMap = resourcesMap
	res.supplier = newSupplier(config, overlay, client, resourcesMap, maxResources)
	res.traders = make([]*trader, minimumNumberOfTraders)
	return res
}

/*============================== DiscoveryInternal Interface =============================== */

func (disc *Discovery) Start() {
	disc.supplier.start()
}

func (disc *Discovery) AddTrader(traderGUID guid.Guid) {
	traderResources, _ := disc.resourcesMap.ResourcesByGuid(traderGUID)
	newTrader := newTrader(disc.config, disc.client, traderGUID, *traderResources)
	disc.traders = append(disc.traders, newTrader)
	log.Printf("[Discovery] New Trader: %s | Resources: %s\n", (&traderGUID).String(), traderResources.ToString())
}

func (disc *Discovery) Find() {
	// TODO
}

func (disc *Discovery) Deploy() {
	// TODO
}

/*============================== DiscoveryExternal Interface ============================== */

func (disc *Discovery) CreateOffer(id int, amount int, suppGUID string, suppIP string) {
	// TODO
}

func (disc *Discovery) RefreshOffer(id int, traderGUID string) {
	// TODO
}

func (disc *Discovery) RemoveOffer(id int, destTraderGUID string, supplierIP string) {
	// TODO
}
