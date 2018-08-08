package random

import (
	"context"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/discovery/backend"
	"github.com/strabox/caravela/node/external"
	"github.com/strabox/caravela/util"
	"sync"
)

type Discovery struct {
	common.NodeComponent // Base component

	config  *configuration.Configuration // System's configurations.
	overlay external.Overlay             // Overlay component.
	client  external.Caravela            // Remote caravela's client.

	maxResources resources.Resources
	resourcesMap *resources.Mapping // GUID<->Resources mapping
	nodeGUID     *guid.GUID

	availableResources resources.Resources
	resourcesMutex     sync.Mutex
}

func NewRandomDiscovery(config *configuration.Configuration, overlay external.Overlay,
	client external.Caravela, resourcesMap *resources.Mapping, maxResources resources.Resources) (backend.Discovery, error) {

	return &Discovery{
		config:  config,
		overlay: overlay,
		client:  client,

		maxResources: maxResources,
		resourcesMap: resourcesMap,
		nodeGUID:     nil,

		availableResources: maxResources,
		resourcesMutex:     sync.Mutex{},
	}, nil
}

// ========================== Internal Services =============================

func (d *Discovery) AddTrader(traderGUID guid.GUID) {
	d.nodeGUID = guid.NewGUIDBytes(traderGUID.Bytes())
	log.Debugf(util.LogTag("RandDisc")+"NEW TRADER GUID: %s", traderGUID.Short())
}

func (d *Discovery) FindOffers(ctx context.Context, targetResources resources.Resources) []types.AvailableOffer {
	// TODO: Makes sense try search in the current node first ?
	resultOffers := make([]types.AvailableOffer, 0)

	if !targetResources.IsValid() { // If the resource combination is not valid we will search for the lowest one
		targetResources = *d.resourcesMap.LowestResources()
	}

	for retry := 0; retry < d.config.RandBackendMaxRetries(); retry++ {
		destinationGUID := guid.NewGUIDRandom()

		reqCtx := context.WithValue(ctx, types.NodeGUIDKey, d.nodeGUID.String())
		nodes, err := d.overlay.Lookup(reqCtx, destinationGUID.Bytes())
		if err != nil {
			continue
		}

		for _, node := range nodes {
			offers, err := d.client.GetOffers(
				ctx,
				&types.Node{GUID: ""},
				&types.Node{IP: node.IP(), GUID: guid.NewGUIDBytes(node.GUID()).String()},
				true,
			)
			if (err == nil) && (len(offers) != 0) {
				for _, offer := range offers {
					tempRes := resources.NewResources(offer.Resources.CPUs, offer.Resources.RAM)
					if tempRes.Contains(targetResources) {
						resultOffers = append(resultOffers, offer)
					}
				}
				if len(resultOffers) != 0 {
					log.Debugf(util.LogTag("RandDisc") + "Offers found")
					return resultOffers // TODO: Request to other nodes (successors for the offers) ??
				}
				continue
			}
		}
	}

	log.Debugf(util.LogTag("RandDisc") + "No offers found")
	return resultOffers
}

func (d *Discovery) ObtainResources(offerID int64, resourcesNecessary resources.Resources) bool {
	d.resourcesMutex.Lock()
	defer d.resourcesMutex.Unlock()

	if d.availableResources.Contains(resourcesNecessary) {
		d.availableResources.Sub(resourcesNecessary)
		return true
	}

	return false
}

func (d *Discovery) ReturnResources(resources resources.Resources) {
	d.resourcesMutex.Lock()
	defer d.resourcesMutex.Unlock()

	d.availableResources.Add(resources)
}

// ======================= External/Remote Services =========================

func (d *Discovery) CreateOffer(_ *types.Node, _ *types.Node, _ *types.Offer) {
	// Do Nothing - Not necessary for this backend.
}

func (d *Discovery) RefreshOffer(_ *types.Node, _ *types.Offer) bool {
	// Do Nothing - Not necessary for this backend.
	return false
}

func (d *Discovery) RemoveOffer(_ *types.Node, _ *types.Node, _ *types.Offer) {
	// Do Nothing - Not necessary for this backend.
}

func (d *Discovery) GetOffers(_ context.Context, _, _ *types.Node, _ bool) []types.AvailableOffer {
	d.resourcesMutex.Lock()
	defer d.resourcesMutex.Unlock()

	if d.availableResources.IsValid() {
		return []types.AvailableOffer{
			{
				SupplierIP: d.config.HostIP(),
				Offer: types.Offer{
					ID:     0,
					Amount: 1,
					Resources: types.Resources{
						CPUs: d.availableResources.CPUs(),
						RAM:  d.availableResources.RAM(),
					},
				},
			},
		}
	}

	return make([]types.AvailableOffer, 0)
}

func (d *Discovery) AdvertiseNeighborOffers(_, _, _ *types.Node) {
	// Do Nothing - Not necessary for this backend.
}

// ============== External/Remote Services (Only Simulation) ================

func (d *Discovery) AvailableResourcesSim() types.Resources {
	d.resourcesMutex.Lock()
	defer d.resourcesMutex.Unlock()
	return types.Resources{
		CPUs: d.availableResources.CPUs(),
		RAM:  d.availableResources.RAM(),
	}
}

func (d *Discovery) MaximumResourcesSim() types.Resources {
	return types.Resources{
		CPUs: d.maxResources.CPUs(),
		RAM:  d.maxResources.RAM(),
	}
}

func (d *Discovery) RefreshOffersSim() {
	// Do Nothing - Not necessary for this backend.
}

func (d *Discovery) SpreadOffersSim() {
	// Do Nothing - Not necessary for this backend.
}

// ===============================================================================
// =							SubComponent Interface                           =
// ===============================================================================

func (d *Discovery) Start() {
	d.Started(d.config.Simulation(), func() {
		// Do Nothing
	})
}

func (d *Discovery) Stop() {
	d.Stopped(func() {
		// Do Nothing
	})
}

func (d *Discovery) IsWorking() bool {
	return d.Working()
}
