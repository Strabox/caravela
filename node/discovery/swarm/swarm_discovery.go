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
	"time"
)

// Zero GUID is the master's GUID (used in simulation only).
const mastersNodeGUID = 0

// Discovery backend is based on a master-slave cluster architecture (Centralized) that simulates the Docker Swarm.
// It is implemented on top of a Chord overlay because it suits better our prototype framework.
// It is NOT DESIGNED to be used in REAL DEPLOYMENT, we only use it in Simulation to compare with our discovery backends.
type Discovery struct {
	common.NodeComponent // Base component.

	// Common fields
	config       *configuration.Configuration // System's configurations.
	overlay      external.Overlay             // Overlay component.
	client       external.Caravela            // Remote caravela's client.
	nodeGUID     *guid.GUID                   // Node's own GUID.
	isMasterNode bool                         // True: if the clusterNode is the master, False: if it is a regular peer.

	// Master clusterNode fields
	clusterNodesByGUID sync.Map       // Map the clusterNode's IP with the clusterNode's structure.
	clusterNodes       []*clusterNode // One clusterNode's structure per clusterNode in the cluster.

	// Peer clusterNode fields
	refreshTicker      <-chan time.Time     //
	containersRunning  int                  //
	maximumResources   *resources.Resources //
	availableResources *resources.Resources //
	resourcesMutex     sync.Mutex           //
}

// NewSwarmResourcesDiscovery creates a resource discovery backend based on the Docker Swarm.
func NewSwarmResourcesDiscovery(config *configuration.Configuration, overlay external.Overlay,
	client external.Caravela, _ *resources.Mapping, maxResources resources.Resources) (backend.Discovery, error) {

	return &Discovery{
		config:       config,
		overlay:      overlay,
		client:       client,
		nodeGUID:     nil,
		isMasterNode: false,

		clusterNodesByGUID: sync.Map{},
		clusterNodes:       make([]*clusterNode, 0),

		refreshTicker:      time.NewTicker(config.RefreshingInterval()).C,
		containersRunning:  0,
		maximumResources:   maxResources.Copy(),
		availableResources: maxResources.Copy(),
		resourcesMutex:     sync.Mutex{},
	}, nil
}

// start starts the discovery backend in the clusterNode.
func (d *Discovery) start() {
	if !d.isMasterNode {
		d.resourcesMutex.Lock()
		defer d.resourcesMutex.Unlock()

		masterNodeIP, masterNodeGUID := d.getMasterNodeIDs()
		usedResources := d.usedResources()
		d.client.CreateOffer(
			context.Background(),
			&types.Node{IP: d.config.HostIP(), GUID: d.nodeGUID.String()},
			&types.Node{IP: masterNodeIP, GUID: masterNodeGUID},
			&types.Offer{
				FreeResources: types.Resources{
					CPUClass: types.CPUClass(d.availableResources.CPUClass()),
					CPUs:     d.availableResources.CPUs(),
					Memory:   d.availableResources.Memory(),
				},
				UsedResources: types.Resources{
					CPUClass: types.CPUClass(usedResources.CPUClass()),
					CPUs:     usedResources.CPUs(),
					Memory:   usedResources.Memory(),
				},
			},
		)
	}

	if !d.config.Simulation() && !d.isMasterNode {
		go func() {
			for {
				select {
				case <-d.refreshTicker:
					masterNodeIP, masterNodeGUID := d.getMasterNodeIDs()

					d.client.RefreshOffer(
						context.Background(),
						&types.Node{IP: d.config.HostIP(), GUID: d.nodeGUID.String()},
						&types.Node{IP: masterNodeIP, GUID: masterNodeGUID},
						&types.Offer{ /* Nothing (only used to simulate real world refreshes in swarm) */ },
					)
				}
			}
		}()
	}
}

// usedResources returns the amount of used resources in this clusterNode (if it is not the master).
func (d *Discovery) usedResources() *resources.Resources {
	usedResources := d.maximumResources.Copy()
	usedResources.Sub(*d.availableResources)
	return usedResources
}

// getMasterNodeIDs returns the IP and GUID of the master clusterNode.
func (d *Discovery) getMasterNodeIDs() (string, string) {
	nodes, _ := d.overlay.Lookup(
		context.Background(),
		guid.NewGUIDInteger(mastersNodeGUID).Bytes(), // Master's clusterNode has GUID 0 (in simulator).
	)

	masterNode := nodes[0]
	return masterNode.IP(), guid.NewGUIDBytes(masterNode.GUID()).String()
}

// ====================== Local Services (Consumed by other Components) ============================

func (d *Discovery) AddTrader(traderGUID guid.GUID) {
	d.nodeGUID = guid.NewGUIDBytes(traderGUID.Bytes())
	d.isMasterNode = d.nodeGUID.Equals(*guid.NewGUIDInteger(mastersNodeGUID))
}

func (d *Discovery) FindOffers(_ context.Context, targetResources resources.Resources) []types.AvailableOffer {
	if d.isMasterNode {
		d.resourcesMutex.Lock()
		defer d.resourcesMutex.Unlock()

		resultOffers := make([]types.AvailableOffer, len(d.clusterNodes))

		// If the resource combination is not valid we will refuse the request.
		if !targetResources.IsValid() {
			return resultOffers
		}

		resultOffersIndex := 0
		for _, clusterNode := range d.clusterNodes {
			// Skip nodes that are smaller than the requested resources.
			if !clusterNode.freeResources.Contains(targetResources) {
				resultOffers = resultOffers[:len(resultOffers)-1]
				continue
			}

			resultOffers[resultOffersIndex] = types.AvailableOffer{
				SupplierIP: clusterNode.ip(),
				Offer: types.Offer{
					FreeResources: types.Resources{
						CPUClass: types.CPUClass(clusterNode.freeResources.CPUClass()),
						CPUs:     clusterNode.freeResources.CPUs(),
						Memory:   clusterNode.freeResources.Memory(),
					},
					UsedResources: types.Resources{
						CPUClass: types.CPUClass(clusterNode.usedResources.CPUClass()),
						CPUs:     clusterNode.usedResources.CPUs(),
						Memory:   clusterNode.usedResources.Memory(),
					},
				},
			}
			resultOffersIndex++
		}

		return resultOffers
	}
	return make([]types.AvailableOffer, 0)
}

func (d *Discovery) ObtainResources(_ int64, resourcesNecessary resources.Resources) bool {
	if !d.isMasterNode {
		d.resourcesMutex.Lock()
		defer d.resourcesMutex.Unlock()

		if d.availableResources.Contains(resourcesNecessary) {
			d.availableResources.Sub(resourcesNecessary)
			d.containersRunning++

			masterNodeIP, masterNodeGUID := d.getMasterNodeIDs()
			usedResources := d.usedResources()
			// Update the resources offered in the master.
			d.client.UpdateOffer(
				context.Background(),
				&types.Node{IP: d.config.HostIP(), GUID: d.nodeGUID.String()},
				&types.Node{IP: masterNodeIP, GUID: masterNodeGUID},
				&types.Offer{
					Amount: d.containersRunning,
					FreeResources: types.Resources{
						CPUClass: types.CPUClass(d.availableResources.CPUClass()),
						CPUs:     d.availableResources.CPUs(),
						Memory:   d.availableResources.Memory(),
					},
					UsedResources: types.Resources{
						CPUClass: types.CPUClass(usedResources.CPUClass()),
						CPUs:     usedResources.CPUs(),
						Memory:   usedResources.Memory(),
					},
				},
			)
			return true
		}
		return false
	}
	return false
}

func (d *Discovery) ReturnResources(releasedResources resources.Resources) {
	if !d.isMasterNode {
		d.resourcesMutex.Lock()
		defer d.resourcesMutex.Unlock()

		d.availableResources.Add(releasedResources)
		d.containersRunning--

		masterNodeIP, masterNodeGUID := d.getMasterNodeIDs()
		usedResources := d.usedResources()
		d.client.UpdateOffer( // Update the resources offered in the master.
			context.Background(),
			&types.Node{IP: d.config.HostIP(), GUID: d.nodeGUID.String()},
			&types.Node{IP: masterNodeIP, GUID: masterNodeGUID},
			&types.Offer{
				Amount: d.containersRunning,
				FreeResources: types.Resources{
					CPUClass: types.CPUClass(d.availableResources.CPUClass()),
					CPUs:     d.availableResources.CPUs(),
					Memory:   d.availableResources.Memory(),
				},
				UsedResources: types.Resources{
					CPUClass: types.CPUClass(usedResources.CPUClass()),
					CPUs:     usedResources.CPUs(),
					Memory:   usedResources.Memory(),
				},
			},
		)
	}
}

// ======================= External Services (Consumed by other Nodes) ==============================

func (d *Discovery) CreateOffer(fromSupp *types.Node, _ *types.Node, offer *types.Offer) {
	if d.isMasterNode {
		d.resourcesMutex.Lock()
		defer d.resourcesMutex.Unlock()

		availableResources := *resources.NewResourcesCPUClass(int(offer.FreeResources.CPUClass), offer.FreeResources.CPUs, offer.FreeResources.Memory)
		usedResources := *resources.NewResourcesCPUClass(int(offer.UsedResources.CPUClass), offer.UsedResources.CPUs, offer.UsedResources.Memory)
		clusterNode := newClusterNode(fromSupp.IP, availableResources, usedResources)

		d.clusterNodes = append(d.clusterNodes, clusterNode)
		d.clusterNodesByGUID.Store(fromSupp.GUID, clusterNode)
	}
}

func (d *Discovery) RefreshOffer(_ *types.Node, _ *types.Offer) bool {
	return true
}

func (d *Discovery) UpdateOffer(fromSupp, _ *types.Node, offer *types.Offer) {
	if d.isMasterNode {
		if nodeStored, exist := d.clusterNodesByGUID.Load(fromSupp.GUID); exist {
			if nodePtr, ok := nodeStored.(*clusterNode); ok {
				nodeFreeUpdatedRes := *resources.NewResourcesCPUClass(int(offer.FreeResources.CPUClass), offer.FreeResources.CPUs, offer.FreeResources.Memory)
				nodeUsedUpdatedRes := *resources.NewResourcesCPUClass(int(offer.UsedResources.CPUClass), offer.UsedResources.CPUs, offer.UsedResources.Memory)

				nodePtr.setFreeResources(nodeFreeUpdatedRes)
				nodePtr.setUsedResources(nodeUsedUpdatedRes)
				nodePtr.setContainerRunning(offer.Amount) // HACK: Careful if we use stack deployments!
			}
		}
	}
}

func (d *Discovery) RemoveOffer(_ *types.Node, _ *types.Node, _ *types.Offer) {
	// Do Nothing - Not necessary for this backend.
}

func (d *Discovery) GetOffers(_ context.Context, _, _ *types.Node, _ bool) []types.AvailableOffer {
	// Do Nothing - Not necessary for this backend.
	return nil
}

func (d *Discovery) AdvertiseNeighborOffers(_, _, _ *types.Node) {
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
		Memory:   d.availableResources.Memory(),
	}
}

// Simulation
func (d *Discovery) MaximumResourcesSim() types.Resources {
	return types.Resources{
		CPUClass: types.CPUClass(d.maximumResources.CPUClass()),
		CPUs:     d.maximumResources.CPUs(),
		Memory:   d.maximumResources.Memory(),
	}
}

// Simulation
func (d *Discovery) RefreshOffersSim() {
	if !d.isMasterNode {
		masterNodeIP, masterNodeGUID := d.getMasterNodeIDs()

		d.client.RefreshOffer(
			context.Background(),
			&types.Node{IP: d.config.HostIP(), GUID: d.nodeGUID.String()},
			&types.Node{IP: masterNodeIP, GUID: masterNodeGUID},
			&types.Offer{ /* Nothing (only used to simulate real world refreshes in swarm) */ },
		)
	}
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
		d.start()
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
