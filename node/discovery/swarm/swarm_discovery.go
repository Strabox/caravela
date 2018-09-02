package swarm

import (
	"context"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/discovery/backend"
	"github.com/strabox/caravela/node/external"
	"sync"
)

type Discovery struct {
	common.NodeComponent // Base component

	config  *configuration.Configuration // System's configurations.
	overlay external.Overlay             // Overlay component.
	client  external.Caravela            // Remote caravela's client.

	nodeGUID *guid.GUID

	//clusterNodesByIP map[string]*node
	clusterNodes []*node // Cluster nodes contains all the nodes.

	maximumResources *resources.Resources //

	availableResources *resources.Resources //
	resourcesMutex     sync.Mutex           //
}

func NewSwarmResourcesDiscovery(config *configuration.Configuration, overlay external.Overlay,
	client external.Caravela, _ *resources.Mapping, maxResources resources.Resources) (backend.Discovery, error) {

	return &Discovery{
		config:   config,
		overlay:  overlay,
		client:   client,
		nodeGUID: nil,
		//clusterNodesByIP: make(map[string]*node),
		clusterNodes:       make([]*node, 0),
		maximumResources:   maxResources.Copy(),
		availableResources: maxResources.Copy(),
		resourcesMutex:     sync.Mutex{},
	}, nil
}

// IsMaster ...
func (d *Discovery) IsMaster() bool {
	return d.nodeGUID.Equals(*guid.NewGUIDInteger(0))
}

// ====================== Local Services (Consumed by other Components) ============================

func (d *Discovery) AddTrader(traderGUID guid.GUID) {
	d.nodeGUID = guid.NewGUIDBytes(traderGUID.Bytes())
}

func (d *Discovery) FindOffers(ctx context.Context, resources resources.Resources) []types.AvailableOffer {
	res := make([]types.AvailableOffer, 0)

	for _, clusterNode := range d.clusterNodes {
		if resources.Contains(clusterNode.availableResources) {
			res = append(res, types.AvailableOffer{
				SupplierIP: clusterNode.ip,
				Offer: types.Offer{
					Resources: types.Resources{
						CPUClass: types.CPUClass(clusterNode.availableResources.CPUClass()),
						CPUs:     clusterNode.availableResources.CPUs(),
						RAM:      clusterNode.availableResources.RAM(),
					},
				},
			})
			return res
		}
	}

	return res
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

func (d *Discovery) ReturnResources(releasedResources resources.Resources) {
	d.resourcesMutex.Lock()
	defer d.resourcesMutex.Unlock()

	d.availableResources.Add(releasedResources)
}

// ======================= External Services (Consumed by other Nodes) ==============================

func (d *Discovery) CreateOffer(fromSupp *types.Node, toTrader *types.Node, offer *types.Offer) {
	d.resourcesMutex.Lock()
	defer d.resourcesMutex.Unlock()

	d.clusterNodes = append(
		d.clusterNodes,
		&node{
			availableResources: *resources.NewResourcesCPUClass(int(offer.Resources.CPUClass), offer.Resources.CPUs, offer.Resources.RAM),
			ip:                 fromSupp.IP,
		})
}

func (d *Discovery) RefreshOffer(fromTrader *types.Node, offer *types.Offer) bool {
	// Do Nothing - Not necessary for this backend.
	// TODO: Refresh offers (pinging each node)
	return false
}

func (d *Discovery) UpdateOffer(fromSupp, toTrader *types.Node, offer *types.Offer) {
	// Do Nothing - Not necessary for this backend.
}

func (d *Discovery) RemoveOffer(fromSupp *types.Node, toTrader *types.Node, offer *types.Offer) {
	// Do Nothing - Not necessary for this backend.
}

func (d *Discovery) GetOffers(ctx context.Context, fromNode, toTrader *types.Node, relay bool) []types.AvailableOffer {
	// Do Nothing - Not necessary for this backend.
	return nil
}

func (d *Discovery) AdvertiseNeighborOffers(fromTrader, toNeighborTrader, traderOffering *types.Node) {
	// Do Nothing - Not necessary for this backend.
}

// ======================= External Services (Consumed during simulation ONLY) =========================

// Simulation
func (d *Discovery) AvailableResourcesSim() types.Resources {
	d.resourcesMutex.Lock()
	defer d.resourcesMutex.Unlock()

	return types.Resources{
		CPUClass: types.CPUClass(d.availableResources.CPUClass()),
		CPUs:     d.availableResources.CPUs(),
		RAM:      d.availableResources.RAM(),
	}
}

// Simulation
func (d *Discovery) MaximumResourcesSim() types.Resources {
	return types.Resources{
		CPUClass: types.CPUClass(d.maximumResources.CPUClass()),
		CPUs:     d.maximumResources.CPUs(),
		RAM:      d.maximumResources.RAM(),
	}
}

// Simulation
func (d *Discovery) RefreshOffersSim() {
	// Do Nothing - Not necessary for this backend.
}

// Simulation
func (d *Discovery) SpreadOffersSim() {
	// Do Nothing - Not necessary for this backend.
}

// ===============================================================================
// =							SubComponent Interface                           =
// ===============================================================================

func (d *Discovery) Start() {
	d.Started(d.config.Simulation(), func() {
		d.resourcesMutex.Lock()
		defer d.resourcesMutex.Unlock()

		if !d.IsMaster() {
			nodes, _ := d.overlay.Lookup(
				context.Background(),
				guid.NewGUIDInteger(0).Bytes(), // Master's node has GUID 0 (in simulator).
			)

			masterNode := nodes[0]
			masterNodeGUIDStr := guid.NewGUIDBytes(masterNode.GUID()).String()
			d.client.CreateOffer(
				context.Background(),
				&types.Node{
					IP:   d.config.HostIP(),
					GUID: d.nodeGUID.String(),
				},
				&types.Node{
					IP:   masterNode.IP(),
					GUID: masterNodeGUIDStr,
				},
				&types.Offer{
					Resources: types.Resources{
						CPUClass: types.CPUClass(d.availableResources.CPUClass()),
						CPUs:     d.availableResources.CPUs(),
						RAM:      d.availableResources.RAM(),
					},
				},
			)
		} else {
			d.clusterNodes = append(
				d.clusterNodes,
				&node{
					availableResources: *resources.NewResourcesCPUClass(int(d.availableResources.CPUClass()), d.availableResources.CPUs(), d.availableResources.RAM()),
					ip:                 d.config.HostIP(),
				})
		}
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
