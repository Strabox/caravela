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

	// Obtain the maximum resources Docker Engine has available
	maxCPUs, maxRAM := dockerClient.GetDockerCPUAndRAM()
	maxAvailableResources := resources.NewResources(maxCPUs, maxRAM)

	// Create Resources Mapping (based on the configurations)
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
func (node *Node) Start(join bool, joinIP string) error {
	var err error

	// Start creating/joining an overlay of CARAVELA nodes
	if join {
		log.Debugln(util.LogTag("Node") + "Joining a overlay...")
		err = node.overlayComp.Join(context.Background(), joinIP, node.config.OverlayPort(), node)
	} else {
		log.Debugln(util.LogTag("Node") + "Creating an overlay...")
		err = node.overlayComp.Create(context.Background(), node)
	}
	if err != nil {
		return err
	}
	log.Debug(util.LogTag("Node") + "Overlay INITIALIZED")

	node.discoveryComp.Start()
	node.containersManagerComp.Start()
	node.schedulerComp.Start()

	err = node.apiServerComp.Start(node) // Start CARAVELA's REST API web server
	if err != nil {
		return err
	}

	log.Debug(util.LogTag("Node") + "Node STARTED")

	if !node.config.Simulation() {
		select {
		case stop := <-node.stopChan: // Block main Goroutine until a stop message is received
			if stop {
				return nil
			}
		}
	}
	return nil
}

// Stop the node's functions.
func (node *Node) Stop(ctx context.Context) {
	log.Debug(util.LogTag("Node") + "STOPPING...")
	node.apiServerComp.Stop()
	log.Debug(util.LogTag("Node") + "-> API SERVER STOPPED")
	node.schedulerComp.Stop()
	log.Debug(util.LogTag("Node") + "-> SCHEDULER STOPPED")
	node.containersManagerComp.Stop()
	log.Debug(util.LogTag("Node") + "-> CONTAINERS MANAGER STOPPED")
	node.discoveryComp.Stop()
	log.Debug(util.LogTag("Node") + "-> DISCOVERY STOPPED")
	node.overlayComp.Leave(context.Background())
	log.Debug(util.LogTag("Node") + "-> OVERLAY STOPPED")
	// Used to make the main goroutine quit and exit the process
	node.stopChan <- true
	log.Debug(util.LogTag("Node") + "-> STOPPED")
}

// Configuration returns the system's configuration of this node.
func (node *Node) Configuration(c context.Context) *configuration.Configuration {
	return node.config
}

// ##############################################################################################
// #									     CLIENT API											#
// ##############################################################################################

// =========================== User Component Interface (USER API) ==============================

func (node *Node) SubmitContainers(ctx context.Context, containerConfigs []types.ContainerConfig) ([]types.ContainerStatus, error) {
	return node.userManagerComp.SubmitContainers(ctx, containerConfigs)
}

func (node *Node) StopContainers(ctx context.Context, containersIDs []string) error {
	return node.userManagerComp.StopContainers(ctx, containersIDs)
}

func (node *Node) ListContainers(_ context.Context) []types.ContainerStatus {
	return node.userManagerComp.ListContainers()
}

// ##############################################################################################
// #								 REMOTE CLIENT API (RPC)  								    #
// ##############################################################################################

// =============================== Overlay Membership Interface =================================

func (node *Node) AddTrader(guidBytes []byte) {
	guidRes := guid.NewGUIDBytes(guidBytes)
	node.discoveryComp.AddTrader(*guidRes)
}

// =============================== Discovery Component Interface =================================

func (node *Node) CreateOffer(ctx context.Context, fromNode *types.Node, toNode *types.Node, offer *types.Offer) {
	if partitionsState := types.SysPartitionsState(ctx); partitionsState != nil {
		node.discoveryComp.UpdatePartitionsState(partitionsState)
	}
	node.discoveryComp.CreateOffer(fromNode, toNode, offer)
}

func (node *Node) RefreshOffer(ctx context.Context, fromTrader *types.Node, offer *types.Offer) bool {
	if partitionsState := types.SysPartitionsState(ctx); partitionsState != nil {
		node.discoveryComp.UpdatePartitionsState(partitionsState)
	}
	return node.discoveryComp.RefreshOffer(fromTrader, offer)
}

func (node *Node) RemoveOffer(ctx context.Context, fromSupp *types.Node, toTrader *types.Node, offer *types.Offer) {
	if partitionsState := types.SysPartitionsState(ctx); partitionsState != nil {
		node.discoveryComp.UpdatePartitionsState(partitionsState)
	}
	node.discoveryComp.RemoveOffer(fromSupp, toTrader, offer)
}

func (node *Node) GetOffers(ctx context.Context, fromNode, toTrader *types.Node, relay bool) []types.AvailableOffer {
	if partitionsState := types.SysPartitionsState(ctx); partitionsState != nil {
		node.discoveryComp.UpdatePartitionsState(partitionsState)
	}
	return node.discoveryComp.GetOffers(ctx, fromNode, toTrader, relay)
}

func (node *Node) AdvertiseOffersNeighbor(ctx context.Context, fromTrader, toNeighborTrader, traderOffering *types.Node) {
	if partitionsState := types.SysPartitionsState(ctx); partitionsState != nil {
		node.discoveryComp.UpdatePartitionsState(partitionsState)
	}
	node.discoveryComp.AdvertiseNeighborOffers(fromTrader, toNeighborTrader, traderOffering)
}

// ================================ Scheduling Component Interface ==============================

func (node *Node) LaunchContainers(ctx context.Context, fromBuyer *types.Node, offer *types.Offer,
	containersConfigs []types.ContainerConfig) ([]types.ContainerStatus, error) {
	return node.schedulerComp.Launch(ctx, fromBuyer, offer, containersConfigs)
}

// ============================== Containers Component Interface ================================

func (node *Node) StopLocalContainer(_ context.Context, containerID string) error {
	return node.containersManagerComp.StopContainer(containerID)
}

// ##############################################################################################
// #									   SIMULATION API									    #
// ##############################################################################################

// =========================== APIs exclusively used in Simulation ==============================

// AvailableResourcesSim returns the current available resources of the node.
// Note: Only available when the node is running in simulation mode.
func (node *Node) AvailableResourcesSim() types.Resources {
	if !node.config.Simulation() {
		panic(errors.New("AvailableResourcesSim request can only be used in simulation mode"))
	}
	return node.discoveryComp.AvailableResourcesSim()
}

// MaximumResourcesSim returns the maximum available resources of the node.
// Note: Only available when the node is running in simulation mode.
func (node *Node) MaximumResourcesSim() types.Resources {
	if !node.config.Simulation() {
		panic(errors.New("MaximumResourcesSim request can only be used in simulation mode"))
	}
	return node.discoveryComp.MaximumResourcesSim()
}

// RefreshOffersSim triggers the inner actions to refresh all the offers that the node is handling.
// Note: Only available when the node is running in simulation mode.
func (node *Node) RefreshOffersSim() {
	if !node.config.Simulation() {
		panic(errors.New("RefreshOffersSim request can only be used in simulation mode"))
	}
	node.discoveryComp.RefreshOffersSim()
}

// SpreadOffersSim triggers the inner actions to spread the offers that the node is handling.
// Note: Only available when the node is running in simulation mode.
func (node *Node) SpreadOffersSim() {
	if !node.config.Simulation() {
		panic(errors.New("SpreadOffersSim request can only be used in simulation mode"))
	}
	node.discoveryComp.SpreadOffersSim()
}
