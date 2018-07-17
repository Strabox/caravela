package discovery

import (
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/api/remote"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/api"
	"github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/discovery/supplier"
	"github.com/strabox/caravela/node/discovery/trader"
	"github.com/strabox/caravela/overlay"
	"github.com/strabox/caravela/util"
	"sync"
)

// Discovery module of a CARAVELA node. It is responsible for dealing with the resource management
// and finding.
type Discovery struct {
	common.NodeComponent // Base component

	config  *configuration.Configuration // System's configurations
	overlay overlay.Overlay              // Node overlay to efficient route messages to specific nodes.
	client  remote.Caravela              // Remote caravela's client

	resourcesMap   *resources.Mapping // GUID<->Resources mapping
	supplier       *supplier.Supplier // Supplier for managing the offers locally and remotely
	virtualTraders sync.Map           // map[string]*trader.Trader // Node can have multiple "virtual" traders in several places of the overlay
}

func NewDiscovery(config *configuration.Configuration, overlay overlay.Overlay,
	client remote.Caravela, resourcesMap *resources.Mapping, maxResources resources.Resources) *Discovery {

	return &Discovery{
		config:  config,
		overlay: overlay,
		client:  client,

		resourcesMap:   resourcesMap,
		supplier:       supplier.NewSupplier(config, overlay, client, resourcesMap, maxResources),
		virtualTraders: sync.Map{},
	}
}

/*============================== DiscoveryInternal Interface =============================== */

// Adds a new local "virtual" trader when the overlay notifies its presence.
func (disc *Discovery) AddTrader(traderGUID guid.GUID) {
	newTrader := trader.NewTrader(disc.config, disc.overlay, disc.client, traderGUID, disc.resourcesMap)
	disc.virtualTraders.Store(traderGUID.String(), newTrader)

	newTrader.Start() // Start the node's trader module.
	newTraderResources, _ := disc.resourcesMap.ResourcesByGUID(traderGUID)
	log.Debugf(util.LogTag("Discovery")+"New Trader: %s | Resources: %s",
		(&traderGUID).String(), newTraderResources.String())
}

func (disc *Discovery) FindOffers(resources resources.Resources) []api.Offer {
	return disc.supplier.FindOffers(resources)
}

func (disc *Discovery) ObtainResources(offerID int64, resourcesNecessary resources.Resources) bool {
	return disc.supplier.ObtainResources(offerID, resourcesNecessary)
}

func (disc *Discovery) ReturnResources(resources resources.Resources) {
	disc.supplier.ReturnResources(resources)
}

/*============================== DiscoveryExternal Interface ============================== */

func (disc *Discovery) CreateOffer(fromSupplierGUID string, fromSupplierIP string, toTraderGUID string,
	id int64, amount int, cpus int, ram int) {

	t, exist := disc.virtualTraders.Load(toTraderGUID)
	toTrader, ok := t.(*trader.Trader)
	if exist && ok {
		toTrader.CreateOffer(id, amount, cpus, ram, fromSupplierGUID, fromSupplierIP)
	}
}

func (disc *Discovery) RefreshOffer(offerID int64, fromTraderGUID string) bool {
	return disc.supplier.RefreshOffer(offerID, fromTraderGUID)
}

func (disc *Discovery) RemoveOffer(fromSupplierIP string, fromSupplierGUID string, toTraderGUID string,
	offerID int64) {

	t, exist := disc.virtualTraders.Load(toTraderGUID)
	toTrader, ok := t.(*trader.Trader)
	if exist && ok {
		toTrader.RemoveOffer(fromSupplierIP, fromSupplierGUID, toTraderGUID, offerID)
	}
}

func (disc *Discovery) GetOffers(toTraderGUID string, relay bool, fromNodeGUID string) []api.Offer {
	t, exist := disc.virtualTraders.Load(toTraderGUID)
	toTrader, ok := t.(*trader.Trader)
	if exist && ok {
		return toTrader.GetOffers(relay, fromNodeGUID)
	} else {
		return nil
	}
}

func (disc *Discovery) AdvertiseNeighborOffers(toTraderGUID string, fromTraderGUID string, traderOfferingIP string,
	traderOfferingGUID string) {

	t, exist := disc.virtualTraders.Load(toTraderGUID)
	toTrader, ok := t.(*trader.Trader)
	if exist && ok {
		toTrader.AdvertiseNeighborOffer(fromTraderGUID, traderOfferingIP, traderOfferingGUID)
	}
}

/*
===============================================================================
							SubComponent Interface
===============================================================================
*/

func (disc *Discovery) Start() {
	disc.Started(disc.config.Simulation(), func() {
		disc.supplier.Start()
	})
}

func (disc *Discovery) Stop() {
	disc.Stopped(func() {
		disc.supplier.Stop()
		disc.virtualTraders.Range(func(_, value interface{}) bool {
			currentTrader, ok := value.(*trader.Trader)
			if ok {
				currentTrader.Stop()
			}
			return true
		})
	})
}

func (disc *Discovery) isWorking() bool {
	return disc.Working()
}
