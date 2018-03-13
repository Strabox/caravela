package discovery

import (
	"fmt"
	"github.com/strabox/caravela/api/client"
	"github.com/strabox/caravela/node/configuration"
	"github.com/strabox/caravela/node/guid"
	"github.com/strabox/caravela/node/resources"
	"github.com/strabox/caravela/overlay"
)

/*
Interface of discovery module for the scheduler
*/
type DiscoveryLocal interface {
	Start()                         // Starts the discovery module operations
	Find()                          // TODO - Request to find a node for a deployment based on resources
	Deploy()                        // TODO - Request to deploy a container in this node
	AddTrader(traderGUID guid.Guid) // Add a new trader (called during overlay bootstrap)
}

/*
Interface of discovery module for other CARAVELA's nodes
*/
type DiscoveryRemote interface {
	Offer(id OfferID, amount int, res resources.Resources, suppGUID guid.Guid, suppIP string)
	Refresh(id OfferID, traderGUID guid.Guid)
}

type Discovery struct {
	resourcesMap *resources.ResourcesMap
	supplier     *supplier
	traders      []*Trader
}

func NewDiscovery(config *configuration.Configuration, overlay overlay.Overlay,
	client client.CaravelaClient, resourcesMap *resources.ResourcesMap,
	maxResources resources.Resources) *Discovery {

	res := &Discovery{}
	res.resourcesMap = resourcesMap
	res.supplier = newSupplier(config, overlay, client, resourcesMap, maxResources)
	res.traders = make([]*Trader, 1)
	return res
}

/*============================== DiscoveryLocal Interface =============================== */

func (disc *Discovery) Start() {
	disc.supplier.Start()
}

func (disc *Discovery) AddTrader(traderGUID guid.Guid) {
	traderResources, _ := disc.resourcesMap.ResourcesByGuid(traderGUID)
	newTrader := NewTrader(traderGUID, *traderResources)
	disc.traders = append(disc.traders, newTrader)
	fmt.Printf("[Discovery] New Trader: %s | Resources: %s\n", (&traderGUID).String(), traderResources.ToString())
}

func (disc *Discovery) Find() {
	// TODO
}

func (disc *Discovery) Deploy() {
	// TODO
}

/*============================== DiscoveryRemote Interface ============================== */

func (disc *Discovery) Offer(id OfferID, amount int, res resources.Resources,
	suppGUID guid.Guid, suppIP string) {
	// TODO
}

func (disc *Discovery) Refresh(id OfferID, traderGUID guid.Guid) {
	// TODO
}
