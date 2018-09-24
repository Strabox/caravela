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

	nodeGUID         *guid.GUID           //
	maximumResources *resources.Resources //
	freeResources    *resources.Resources //
	resourcesMutex   sync.Mutex           //
}

func NewRandomDiscovery(_ common.Node, config *configuration.Configuration, overlay external.Overlay,
	client external.Caravela, _ *resources.Mapping, maxResources resources.Resources) (backend.Discovery, error) {

	return &Discovery{
		config:  config,
		overlay: overlay,
		client:  client,

		maximumResources: maxResources.Copy(),
		nodeGUID:         nil,

		freeResources:  maxResources.Copy(),
		resourcesMutex: sync.Mutex{},
	}, nil
}

// ========================== Internal Services =============================

func (d *Discovery) AddTrader(traderGUID guid.GUID) {
	d.nodeGUID = guid.NewGUIDBytes(traderGUID.Bytes())
	log.Debugf(util.LogTag("RandDisc")+"NEW TRADER GUID: %s", traderGUID.Short())
}

func (d *Discovery) FindOffers(ctx context.Context, targetResources resources.Resources) []types.AvailableOffer {
	resultOffers := make([]types.AvailableOffer, 0)

	if !targetResources.IsValid() { // If the resource combination is not valid we will refuse the request.
		return resultOffers
	}

	for retry := 0; retry < d.config.RandBackendMaxRetries(); retry++ {
		destinationGUID := guid.NewGUIDRandom()

		ctx := context.WithValue(ctx, types.NodeGUIDKey, d.nodeGUID.String())
		nodes, err := d.overlay.Lookup(ctx, destinationGUID.Bytes())
		if err != nil {
			continue
		}

		for _, node := range nodes {
			offers, err := d.client.GetOffers(
				ctx,
				&types.Node{},
				&types.Node{IP: node.IP(), GUID: guid.NewGUIDBytes(node.GUID()).String()},
				true,
			)
			if err == nil && len(offers) != 0 {
				for _, offer := range offers {
					tempRes := resources.NewResourcesCPUClass(int(offer.FreeResources.CPUClass), offer.FreeResources.CPUs, offer.FreeResources.Memory)
					if tempRes.Contains(targetResources) {
						resultOffers = append(resultOffers, offer)
					}
				}
				if len(resultOffers) != 0 {
					log.Debugf(util.LogTag("RandDisc") + "Offers found")
					return resultOffers
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

	if d.freeResources.Contains(resourcesNecessary) {
		d.freeResources.Sub(resourcesNecessary)
		return true
	}

	return false
}

func (d *Discovery) ReturnResources(releasedResources resources.Resources) {
	d.resourcesMutex.Lock()
	defer d.resourcesMutex.Unlock()

	d.freeResources.Add(releasedResources)
}

// ======================= External/Remote Services =========================

func (d *Discovery) CreateOffer(_ *types.Node, _ *types.Node, _ *types.Offer) {
	// Do Nothing - Not necessary for this backend.
}

func (d *Discovery) RefreshOffer(_ *types.Node, _ *types.Offer) bool {
	// Do Nothing - Not necessary for this backend.
	return false
}

func (d *Discovery) UpdateOffer(_, _ *types.Node, _ *types.Offer) {
	// Do Nothing - Not necessary for this backend.
}

func (d *Discovery) RemoveOffer(_ *types.Node, _ *types.Node, _ *types.Offer) {
	// Do Nothing - Not necessary for this backend.
}

func (d *Discovery) GetOffers(_ context.Context, _, _ *types.Node, _ bool) []types.AvailableOffer {
	d.resourcesMutex.Lock()
	defer d.resourcesMutex.Unlock()

	if d.freeResources.IsValid() {
		usedResources := d.maximumResources.Copy()
		usedResources.Sub(*d.freeResources)
		return []types.AvailableOffer{
			{
				SupplierIP: d.config.HostIP(),
				Offer: types.Offer{
					ID:     0,
					Amount: 1,
					FreeResources: types.Resources{
						CPUClass: types.CPUClass(d.freeResources.CPUClass()),
						CPUs:     d.freeResources.CPUs(),
						Memory:   d.freeResources.Memory(),
					},
					UsedResources: types.Resources{
						CPUClass: types.CPUClass(usedResources.CPUClass()),
						CPUs:     usedResources.CPUs(),
						Memory:   usedResources.Memory(),
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

func (d *Discovery) NodeInformationSim() (types.Resources, types.Resources, int) {
	d.resourcesMutex.Lock()
	defer d.resourcesMutex.Unlock()
	return types.Resources{
			CPUClass: types.CPUClass(d.freeResources.CPUClass()),
			CPUs:     d.freeResources.CPUs(),
			Memory:   d.freeResources.Memory(),
		},
		types.Resources{
			CPUClass: types.CPUClass(d.freeResources.CPUClass()),
			CPUs:     d.maximumResources.CPUs(),
			Memory:   d.maximumResources.Memory(),
		},
		0
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
