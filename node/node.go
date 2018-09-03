/*
Node package contains the main logic for a CARAVELA's node. It represents a CARAVELA's node it has a one-to-one
relation with the number of daemons/machine in the system.
It is the facade for all the functionality exposing it.
*/
package node

import (
	"context"
	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/strabox/caravela/api"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/containers"
	"github.com/strabox/caravela/node/discovery"
	"github.com/strabox/caravela/node/discovery/backend"
	"github.com/strabox/caravela/node/discovery/offering/partitions"
	"github.com/strabox/caravela/node/external"
	"github.com/strabox/caravela/node/scheduler"
	"github.com/strabox/caravela/node/user"
	"github.com/strabox/caravela/util"
)

// Node is the top level entry structure, facade, for all the functionality of a CARAVELA's node.
type Node struct {
	apiServerComp         api.Server           // API server component.
	discoveryComp         backend.Discovery    // Discovery component.
	schedulerComp         *scheduler.Scheduler // Scheduler component.
	containersManagerComp *containers.Manager  // Container's Manager component.
	userManagerComp       *user.Manager        // User's Manager component.
	overlayComp           external.Overlay     // Overlay component.

	config   *configuration.Configuration // System's configurations.
	stopChan chan bool                    // Channel to stop the node functions.
}

// NewNode creates a Node object that contains all the functionality of a CARAVELA's node.
func NewNode(config *configuration.Configuration, overlay external.Overlay, caravelaCli external.Caravela,
	dockerClient external.DockerClient, apiServer api.Server) *Node {

	cpuClass, maxCPUs, maxRAM := dockerClient.GetDockerEngineTotalResources() // Obtain the maximum resources Docker Engine has available
	maxCPUs = int((float64(maxCPUs) * float64(config.CPUOvercommit())) / 100) // Apply CPU Overcommit factor
	maxRAM = int((float64(maxRAM) * float64(config.RAMOvercommit())) / 100)   // Apply RAM Overcommit factor
	CPUSlices := maxCPUs * config.CPUSlices()                                 // Calculate the CPU slices
	maxAvailableResources := resources.NewResourcesCPUClass(cpuClass, CPUSlices, maxRAM)

	// Create FreeResources Mapping (based on the configurations)
	resourcesMap := resources.NewResourcesMap(resources.ObtainConfiguredPartitions(config.ResourcesPartitions()))

	// Create all the internal components

	discoveryComp := discovery.CreateDiscoveryBackend(config, overlay, caravelaCli, resourcesMap, *maxAvailableResources)

	containersManagerComp := containers.NewManager(config, dockerClient, discoveryComp)

	schedulerComp := scheduler.NewScheduler(config, discoveryComp, containersManagerComp, caravelaCli)

	userManagerComp := user.NewManager(config, schedulerComp, caravelaCli, *resourcesMap.LowestResources())

	return &Node{
		apiServerComp:         apiServer,
		discoveryComp:         discoveryComp,
		schedulerComp:         schedulerComp,
		containersManagerComp: containersManagerComp,
		userManagerComp:       userManagerComp,
		overlayComp:           overlay,

		config:   config,
		stopChan: make(chan bool),
	}
}

// Start the node's functions. If the node is joining an instance of CARAVELA's it is called with join
// as true and the joinIP contains the IP address of a node that already belongs to the CARAVELA's instance.
func (n *Node) Start(join bool, joinIP string) error {
	var err error

	// Start creating/joining an overlay of CARAVELA nodes
	if join {
		log.Debugln(util.LogTag("Node") + "Joining a overlay...")
		err = n.overlayComp.Join(context.Background(), joinIP, n.config.OverlayPort(), n)
	} else {
		log.Debugln(util.LogTag("Node") + "Creating an overlay...")
		err = n.overlayComp.Create(context.Background(), n)
	}
	if err != nil {
		return err
	}
	log.Debug(util.LogTag("Node") + "Overlay INITIALIZED")

	n.discoveryComp.Start()
	n.containersManagerComp.Start()
	n.schedulerComp.Start()

	err = n.apiServerComp.Start(n) // Start CARAVELA's REST API web server
	if err != nil {
		return err
	}

	log.Debug(util.LogTag("Node") + "Node STARTED")

	if !n.config.Simulation() {
		select {
		case stop := <-n.stopChan: // Block main Goroutine until a stop message is received
			if stop {
				return nil
			}
		}
	}
	return nil
}

// Stop the node's functions.
func (n *Node) Stop(ctx context.Context) {
	log.Debug(util.LogTag("Node") + "STOPPING...")
	n.apiServerComp.Stop()
	log.Debug(util.LogTag("Node") + "-> API SERVER STOPPED")
	n.schedulerComp.Stop()
	log.Debug(util.LogTag("Node") + "-> SCHEDULER STOPPED")
	n.containersManagerComp.Stop()
	log.Debug(util.LogTag("Node") + "-> CONTAINERS MANAGER STOPPED")
	n.discoveryComp.Stop()
	log.Debug(util.LogTag("Node") + "-> DISCOVERY STOPPED")
	n.overlayComp.Leave(context.Background())
	log.Debug(util.LogTag("Node") + "-> OVERLAY STOPPED")
	// Used to make the main goroutine quit and exit the process
	n.stopChan <- true
	log.Debug(util.LogTag("Node") + "-> STOPPED")
}

// Configuration returns the system's configuration of this node.
func (n *Node) Configuration(c context.Context) *configuration.Configuration {
	return n.config
}

// ##############################################################################################
// #									     CLIENT API											#
// ##############################################################################################

// =========================== User Component Interface (USER API) ==============================

func (n *Node) SubmitContainers(ctx context.Context, containerConfigs []types.ContainerConfig) ([]types.ContainerStatus, error) {
	return n.userManagerComp.SubmitContainers(ctx, containerConfigs)
}

func (n *Node) StopContainers(ctx context.Context, containersIDs []string) error {
	return n.userManagerComp.StopContainers(ctx, containersIDs)
}

func (n *Node) ListContainers(_ context.Context) []types.ContainerStatus {
	return n.userManagerComp.ListContainers()
}

// ##############################################################################################
// #								 REMOTE CLIENT API (RPC)  								    #
// ##############################################################################################

// =============================== Overlay Membership Interface =================================

func (n *Node) AddTrader(guidBytes []byte) {
	guidRes := guid.NewGUIDBytes(guidBytes)
	n.discoveryComp.AddTrader(*guidRes)
}

// =============================== Discovery Component Interface =================================

func (n *Node) CreateOffer(ctx context.Context, fromNode *types.Node, toNode *types.Node, offer *types.Offer) {
	if partitionsState := types.SysPartitionsState(ctx); partitionsState != nil && n.config.SpreadPartitionsState() {
		partitions.GlobalState.MergePartitionsState(partitionsState)
	}
	n.discoveryComp.CreateOffer(fromNode, toNode, offer)
}

func (n *Node) UpdateOffer(ctx context.Context, fromSupplier, toTrader *types.Node, offer *types.Offer) {
	if partitionsState := types.SysPartitionsState(ctx); partitionsState != nil && n.config.SpreadPartitionsState() {
		partitions.GlobalState.MergePartitionsState(partitionsState)
	}
	n.discoveryComp.UpdateOffer(fromSupplier, toTrader, offer)
}

func (n *Node) RefreshOffer(ctx context.Context, fromTrader *types.Node, offer *types.Offer) bool {
	if partitionsState := types.SysPartitionsState(ctx); partitionsState != nil && n.config.SpreadPartitionsState() {
		partitions.GlobalState.MergePartitionsState(partitionsState)
	}
	return n.discoveryComp.RefreshOffer(fromTrader, offer)
}

func (n *Node) RemoveOffer(ctx context.Context, fromSupp *types.Node, toTrader *types.Node, offer *types.Offer) {
	if partitionsState := types.SysPartitionsState(ctx); partitionsState != nil && n.config.SpreadPartitionsState() {
		partitions.GlobalState.MergePartitionsState(partitionsState)
	}
	n.discoveryComp.RemoveOffer(fromSupp, toTrader, offer)
}

func (n *Node) GetOffers(ctx context.Context, fromNode, toTrader *types.Node, relay bool) []types.AvailableOffer {
	if partitionsState := types.SysPartitionsState(ctx); partitionsState != nil && n.config.SpreadPartitionsState() {
		partitions.GlobalState.MergePartitionsState(partitionsState)
	}
	return n.discoveryComp.GetOffers(ctx, fromNode, toTrader, relay)
}

func (n *Node) AdvertiseOffersNeighbor(ctx context.Context, fromTrader, toNeighborTrader, traderOffering *types.Node) {
	if partitionsState := types.SysPartitionsState(ctx); partitionsState != nil && n.config.SpreadPartitionsState() {
		partitions.GlobalState.MergePartitionsState(partitionsState)
	}
	n.discoveryComp.AdvertiseNeighborOffers(fromTrader, toNeighborTrader, traderOffering)
}

// ================================ Scheduling Component Interface ==============================

func (n *Node) LaunchContainers(ctx context.Context, fromBuyer *types.Node, offer *types.Offer,
	containersConfigs []types.ContainerConfig) ([]types.ContainerStatus, error) {
	if partitionsState := types.SysPartitionsState(ctx); partitionsState != nil && n.config.SpreadPartitionsState() {
		partitions.GlobalState.MergePartitionsState(partitionsState)
	}
	return n.schedulerComp.Launch(ctx, fromBuyer, offer, containersConfigs)
}

// ============================== Containers Component Interface ================================

func (n *Node) StopLocalContainer(ctx context.Context, containerID string) error {
	if partitionsState := types.SysPartitionsState(ctx); partitionsState != nil && n.config.SpreadPartitionsState() {
		partitions.GlobalState.MergePartitionsState(partitionsState)
	}
	return n.containersManagerComp.StopContainer(containerID)
}

// ##############################################################################################
// #									   SIMULATION API									    #
// ##############################################################################################

// =========================== APIs exclusively used in Simulation ==============================

// AvailableResourcesSim returns the current available resources of the node.
// Note: Only available when the node is running in simulation mode.
func (n *Node) AvailableResourcesSim() types.Resources {
	if !n.config.Simulation() {
		panic(errors.New("AvailableResourcesSim request can only be used in Simulation Mode"))
	}
	return n.discoveryComp.AvailableResourcesSim()
}

// MaximumResourcesSim returns the maximum available resources of the node.
// Note: Only available when the node is running in simulation mode.
func (n *Node) MaximumResourcesSim() types.Resources {
	if !n.config.Simulation() {
		panic(errors.New("MaximumResourcesSim request can only be used in Simulation Mode"))
	}
	return n.discoveryComp.MaximumResourcesSim()
}

// RefreshOffersSim triggers the inner actions to refresh all the offers that the node is handling.
// Note: Only available when the node is running in simulation mode.
func (n *Node) RefreshOffersSim() {
	if !n.config.Simulation() {
		panic(errors.New("RefreshOffersSim request can only be used in Simulation Mode"))
	}
	n.discoveryComp.RefreshOffersSim()
}

// SpreadOffersSim triggers the inner actions to spread the offers that the node is handling.
// Note: Only available when the node is running in simulation mode.
func (n *Node) SpreadOffersSim() {
	if !n.config.Simulation() {
		panic(errors.New("SpreadOffersSim request can only be used in Simulation Mode"))
	}
	n.discoveryComp.SpreadOffersSim()
}
