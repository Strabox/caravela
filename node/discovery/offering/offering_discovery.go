package offering

import (
	"context"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/discovery/backend"
	"github.com/strabox/caravela/node/discovery/offering/supplier"
	"github.com/strabox/caravela/node/discovery/offering/trader"
	"github.com/strabox/caravela/node/external"
	"github.com/strabox/caravela/util"
	"sync"
)

// Discovery is responsible for dealing with the resource management local and remote.
// It allows the other components to use its services.
type Discovery struct {
	common.NodeComponent // Base component

	config  *configuration.Configuration // System's configurations.
	overlay external.Overlay             // Overlay component.
	client  external.Caravela            // Remote caravela's client.

	resourcesMap *resources.Mapping // GUID<->FreeResources mapping
	supplier     *supplier.Supplier // Supplier for managing the offers locally and remotely
	traders      sync.Map           // Node can have multiple "virtual" traders in several places of the overlay
}

func NewOfferingDiscovery(node common.Node, config *configuration.Configuration, overlay external.Overlay,
	client external.Caravela, resourcesMap *resources.Mapping, maxResources resources.Resources) (backend.Discovery, error) {

	return &Discovery{
		config:  config,
		overlay: overlay,
		client:  client,

		resourcesMap: resourcesMap,
		supplier:     supplier.NewSupplier(node, config, overlay, client, resourcesMap, maxResources),
		traders:      sync.Map{},
	}, nil
}

// ====================== Local Services (Consumed by other Components) ============================

// Adds a new local "virtual" trader when the overlay notifies its presence.
func (d *Discovery) AddTrader(traderGUID guid.GUID) {
	d.supplier.SetNodeGUID(traderGUID)

	newTrader := trader.NewTrader(d.config, d.overlay, d.client, traderGUID, d.resourcesMap)
	d.traders.Store(traderGUID.String(), newTrader)

	newTrader.Start() // Start the node's trader module.
	newTraderResources := d.resourcesMap.ResourcesByGUID(traderGUID)
	log.Debugf(util.LogTag("DISCOVERY")+"NEW TRADER GUID: %s, Res: %s", traderGUID.Short(), newTraderResources.String())
}

func (d *Discovery) FindOffers(ctx context.Context, resources resources.Resources) []types.AvailableOffer {
	return d.supplier.FindOffers(ctx, resources)
}

func (d *Discovery) ObtainResources(offerID int64, resourcesNecessary resources.Resources) bool {
	return d.supplier.ObtainResources(offerID, resourcesNecessary)
}

func (d *Discovery) ReturnResources(resources resources.Resources) {
	d.supplier.ReturnResources(resources)
}

// ======================= External Services (Consumed by other Nodes) ==============================

func (d *Discovery) CreateOffer(fromSupp *types.Node, toTrader *types.Node, offer *types.Offer) {
	t, exist := d.traders.Load(toTrader.GUID)
	targetTrader, ok := t.(*trader.Trader)
	if exist && ok {
		targetTrader.CreateOffer(fromSupp, offer)
	}
}

func (d *Discovery) RefreshOffer(fromTrader *types.Node, offer *types.Offer) bool {
	return d.supplier.RefreshOffer(fromTrader, offer)
}

func (d *Discovery) UpdateOffer(fromSupp, toTrader *types.Node, offer *types.Offer) {
	t, exist := d.traders.Load(toTrader.GUID)
	targetTrader, ok := t.(*trader.Trader)
	if exist && ok {
		targetTrader.UpdateOffer(fromSupp, offer)
	}
}

func (d *Discovery) RemoveOffer(fromSupp *types.Node, toTrader *types.Node, offer *types.Offer) {
	t, exist := d.traders.Load(toTrader.GUID)
	targetTrader, ok := t.(*trader.Trader)
	if exist && ok {
		targetTrader.RemoveOffer(fromSupp, offer)
	}
}

func (d *Discovery) GetOffers(ctx context.Context, fromNode, toTrader *types.Node, relay bool) []types.AvailableOffer {
	t, exist := d.traders.Load(toTrader.GUID)
	targetTrader, ok := t.(*trader.Trader)
	if exist && ok {
		return targetTrader.GetOffers(ctx, fromNode, relay)
	} else {
		return nil
	}
}

func (d *Discovery) AdvertiseNeighborOffers(fromTrader, toNeighborTrader, traderOffering *types.Node) {
	t, exist := d.traders.Load(toNeighborTrader.GUID)
	targetTrader, ok := t.(*trader.Trader)
	if exist && ok {
		targetTrader.AdvertiseNeighborOffer(fromTrader, traderOffering)
	}
}

// ======================= External Services (Consumed during simulation ONLY) =========================

// Simulation
func (d *Discovery) AvailableResourcesSim() types.Resources {
	return d.supplier.AvailableResources()
}

// Simulation
func (d *Discovery) MaximumResourcesSim() types.Resources {
	return d.supplier.MaximumResources()
}

// Simulation
func (d *Discovery) RefreshOffersSim() {
	d.traders.Range(func(key, value interface{}) bool {
		currentTrader, ok := value.(*trader.Trader)
		if ok {
			currentTrader.RefreshOffersSim()
		}
		return true
	})
}

// Simulation
func (d *Discovery) SpreadOffersSim() {
	d.traders.Range(func(key, value interface{}) bool {
		currentTrader, ok := value.(*trader.Trader)
		if ok {
			currentTrader.SpreadOffersSim()
		}
		return true
	})
}

// ===============================================================================
// =							SubComponent Interface                           =
// ===============================================================================

func (d *Discovery) Start() {
	d.Started(d.config.Simulation(), func() {
		d.supplier.Start()
	})
}

func (d *Discovery) Stop() {
	d.Stopped(func() {
		d.supplier.Stop()
		d.traders.Range(func(_, value interface{}) bool {
			currentTrader, ok := value.(*trader.Trader)
			if ok {
				currentTrader.Stop()
			}
			return true
		})
	})
}

func (d *Discovery) IsWorking() bool {
	return d.Working()
}
