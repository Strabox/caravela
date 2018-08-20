package soffer

import (
	"context"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/discovery/backend"
	"github.com/strabox/caravela/node/discovery/soffer/supplier"
	"github.com/strabox/caravela/node/discovery/soffer/trader"
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

	resourcesMap   *resources.Mapping // GUID<->Resources mapping
	supplier       *supplier.Supplier // Supplier for managing the offers locally and remotely
	virtualTraders sync.Map           // Node can have multiple "virtual" traders in several places of the overlay
}

func NewSOfferDiscovery(config *configuration.Configuration, overlay external.Overlay,
	client external.Caravela, resourcesMap *resources.Mapping, maxResources resources.Resources) (backend.Discovery, error) {

	return &Discovery{
		config:  config,
		overlay: overlay,
		client:  client,

		resourcesMap:   resourcesMap,
		supplier:       supplier.NewSupplier(config, overlay, client, resourcesMap, maxResources),
		virtualTraders: sync.Map{},
	}, nil
}

// ====================== Local Services (Consumed by other Components) ============================

// Adds a new local "virtual" trader when the overlay notifies its presence.
func (disc *Discovery) AddTrader(traderGUID guid.GUID) {
	disc.supplier.SetNodeGUID(traderGUID)

	newTrader := trader.NewTrader(disc.config, disc.overlay, disc.client, traderGUID, disc.resourcesMap)
	disc.virtualTraders.Store(traderGUID.String(), newTrader)

	newTrader.Start() // Start the node's trader module.
	newTraderResources := disc.resourcesMap.ResourcesByGUID(traderGUID)
	log.Debugf(util.LogTag("DISCOVERY")+"NEW TRADER GUID: %s, Res: %s", traderGUID.Short(), newTraderResources.String())
}

func (disc *Discovery) FindOffers(ctx context.Context, resources resources.Resources) []types.AvailableOffer {
	return disc.supplier.FindOffers(ctx, resources)
}

func (disc *Discovery) ObtainResources(offerID int64, resourcesNecessary resources.Resources) bool {
	return disc.supplier.ObtainResources(offerID, resourcesNecessary)
}

func (disc *Discovery) ReturnResources(resources resources.Resources) {
	disc.supplier.ReturnResources(resources)
}

// ======================= External Services (Consumed by other Nodes) ==============================

func (disc *Discovery) CreateOffer(fromNode *types.Node, toNode *types.Node, offer *types.Offer) {
	t, exist := disc.virtualTraders.Load(toNode.GUID)
	targetTrader, ok := t.(*trader.Trader)
	if exist && ok {
		targetTrader.CreateOffer(fromNode, offer)
	}
}

func (disc *Discovery) RefreshOffer(fromTrader *types.Node, offer *types.Offer) bool {
	return disc.supplier.RefreshOffer(fromTrader, offer)
}

func (disc *Discovery) RemoveOffer(fromSupp *types.Node, toTrader *types.Node, offer *types.Offer) {
	t, exist := disc.virtualTraders.Load(toTrader.GUID)
	targetTrader, ok := t.(*trader.Trader)
	if exist && ok {
		targetTrader.RemoveOffer(fromSupp, offer)
	}
}

func (disc *Discovery) GetOffers(ctx context.Context, fromNode, toTrader *types.Node, relay bool) []types.AvailableOffer {
	t, exist := disc.virtualTraders.Load(toTrader.GUID)
	targetTrader, ok := t.(*trader.Trader)
	if exist && ok {
		return targetTrader.GetOffers(ctx, fromNode, relay)
	} else {
		return nil
	}
}

func (disc *Discovery) AdvertiseNeighborOffers(fromTrader, toNeighborTrader, traderOffering *types.Node) {
	t, exist := disc.virtualTraders.Load(toNeighborTrader.GUID)
	targetTrader, ok := t.(*trader.Trader)
	if exist && ok {
		targetTrader.AdvertiseNeighborOffer(fromTrader, traderOffering)
	}
}

// ======================= External Services (Consumed during simulation ONLY) =========================

// Simulation
func (disc *Discovery) AvailableResourcesSim() types.Resources {
	return disc.supplier.AvailableResources()
}

// Simulation
func (disc *Discovery) MaximumResourcesSim() types.Resources {
	return disc.supplier.MaximumResources()
}

// Simulation
func (disc *Discovery) RefreshOffersSim() {
	disc.virtualTraders.Range(func(key, value interface{}) bool {
		currentTrader, ok := value.(*trader.Trader)
		if ok {
			currentTrader.RefreshOffersSim()
		}
		return true
	})
}

// Simulation
func (disc *Discovery) SpreadOffersSim() {
	disc.virtualTraders.Range(func(key, value interface{}) bool {
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

func (disc *Discovery) IsWorking() bool {
	return disc.Working()
}
