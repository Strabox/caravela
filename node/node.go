package node

import (
	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/strabox/caravela/api"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/containers"
	"github.com/strabox/caravela/node/discovery"
	"github.com/strabox/caravela/node/external"
	"github.com/strabox/caravela/node/scheduler"
	"github.com/strabox/caravela/node/user"
	"github.com/strabox/caravela/util"
)

// Node is the top level entry structure, facade, for all the functionality of a CARAVELA's node.
type Node struct {
	config   *configuration.Configuration // System's configuration
	stopChan chan bool                    // Channel to stop the node functions

	apiServer         api.Server           // API server component
	discovery         *discovery.Discovery // Discovery component
	scheduler         *scheduler.Scheduler // Scheduler component
	containersManager *containers.Manager  // Containers Manager component
	userManager       *user.Manager        // User Manager component
	overlay           external.Overlay     // Overlay component
}

// NewNode creates a Node object that contains all the functionality of a CARAVELA's node.
func NewNode(config *configuration.Configuration, overlay external.Overlay, caravelaCli external.Caravela,
	dockerClient external.DockerClient, apiServer api.Server) *Node {

	// Obtain the maximum resources Docker Engine has available
	maxCPUs, maxRAM := dockerClient.GetDockerCPUAndRAM()
	maxAvailableResources := resources.NewResources(maxCPUs, maxRAM)

	// Create Resources Mapping (based on the configurations)
	resourcesMap := resources.NewResourcesMap(resources.GetCpuCoresPartitions(config.CPUCoresPartitions()),
		resources.GetRamPartitions(config.RAMPartitions()))

	// Create all the internal components

	discoveryComp := discovery.NewDiscovery(config, overlay, caravelaCli, resourcesMap, *maxAvailableResources)

	containersManagerComp := containers.NewManager(config, dockerClient, discoveryComp)

	schedulerComp := scheduler.NewScheduler(config, discoveryComp, containersManagerComp, caravelaCli)

	userManagerComp := user.NewManager(config, schedulerComp, caravelaCli)

	return &Node{
		config:   config,
		stopChan: make(chan bool),

		apiServer:         apiServer,
		overlay:           overlay,
		discovery:         discoveryComp,
		containersManager: containersManagerComp,
		scheduler:         schedulerComp,
		userManager:       userManagerComp,
	}
}

// Configuration returns the system's configuration of this CARAVELA's node.
func (node *Node) Configuration() *configuration.Configuration {
	return node.config
}

// ================================ SubComponent Interface ================================

func (node *Node) Start(join bool, joinIP string) error {
	var err error

	// Start creating/joining an overlay of CARAVELA nodes
	if join {
		log.Debugln(util.LogTag("Node") + "Joining a overlay ...")
		err = node.overlay.Join(joinIP, node.config.OverlayPort(), node)
	} else {
		log.Debugln(util.LogTag("Node") + "Creating an overlay ...")
		err = node.overlay.Create(node)
	}
	if err != nil {
		return err
	}
	log.Debug(util.LogTag("Node") + "Overlay INITIALIZED")

	node.discovery.Start()
	node.containersManager.Start()
	node.scheduler.Start()

	err = node.apiServer.Start(node) // Start CARAVELA REST API web server
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

func (node *Node) Stop() {
	log.Debug(util.LogTag("Node") + "STOPPING...")
	node.apiServer.Stop()
	log.Debug(util.LogTag("Node") + "-> API SERVER STOPPED")
	node.scheduler.Stop()
	log.Debug(util.LogTag("Node") + "-> SCHEDULER STOPPED")
	node.containersManager.Stop()
	log.Debug(util.LogTag("Node") + "-> CONTAINERS MANAGER STOPPED")
	node.discovery.Stop()
	log.Debug(util.LogTag("Node") + "-> DISCOVERY STOPPED")
	node.overlay.Leave()
	log.Debug(util.LogTag("Node") + "-> OVERLAY STOPPED")
	// Used to make the main goroutine quit and exit the process
	node.stopChan <- true
	log.Debug(util.LogTag("Node") + "-> STOPPED")
}

// ##############################################################################################
// #								   REMOTE CLIENT API (RPC)  								#
// ##############################################################################################

// ================================= Overlay Membership Interface ===============================

func (node *Node) AddTrader(guidBytes []byte) {
	guidRes := guid.NewGUIDBytes(guidBytes)
	node.discovery.AddTrader(*guidRes)
}

// ================================== Discovery Component Interface =============================

func (node *Node) CreateOffer(fromNode *types.Node, toNode *types.Node, offer *types.Offer) {
	node.discovery.CreateOffer(fromNode, toNode, offer)
}

func (node *Node) RefreshOffer(fromTrader *types.Node, offer *types.Offer) bool {
	return node.discovery.RefreshOffer(fromTrader, offer)
}

func (node *Node) RemoveOffer(fromSupp *types.Node, toTrader *types.Node, offer *types.Offer) {
	node.discovery.RemoveOffer(fromSupp, toTrader, offer)
}

func (node *Node) GetOffers(fromNode, toTrader *types.Node, relay bool) []types.AvailableOffer {
	return node.discovery.GetOffers(fromNode, toTrader, relay)
}

func (node *Node) AdvertiseOffersNeighbor(fromTrader, toNeighborTrader, traderOffering *types.Node) {
	node.discovery.AdvertiseNeighborOffers(fromTrader, toNeighborTrader, traderOffering)
}

// ================================ Scheduling Component Interface ==============================

func (node *Node) LaunchContainers(fromBuyer *types.Node, offer *types.Offer,
	containerConfig *types.ContainerConfig) (*types.ContainerStatus, error) {
	return node.scheduler.Launch(fromBuyer, offer, containerConfig)
}

// ============================== Containers Component Interface ================================

func (node *Node) StopLocalContainer(containerID string) error {
	return node.containersManager.StopContainer(containerID)
}

// ##############################################################################################
// #									     CLIENT API											#
// ##############################################################################################

// =========================== User Component Interface (USER API) ==============================

func (node *Node) SubmitContainers(containerImageKey string, portMappings []types.PortMapping,
	containerArgs []string, cpus int, ram int) error {
	return node.userManager.SubmitContainers(containerImageKey, portMappings, containerArgs, cpus, ram)
}

func (node *Node) StopContainers(containersIDs []string) error {
	return node.userManager.StopContainers(containersIDs)
}

func (node *Node) ListContainers() []types.ContainerStatus {
	return node.userManager.ListContainers()
}

// ##############################################################################################
// #									   SIMULATION API									    #
// ##############################################################################################

// =========================== APIs exclusively used in Simulation ==============================

func (node *Node) AvailableResourcesSim() types.Resources {
	if !node.config.Simulation() {
		panic(errors.New("AvailableResourcesSim request can only be used in simulation mode"))
	}
	return node.discovery.AvailableResourcesSim()
}

func (node *Node) MaximumResourcesSim() types.Resources {
	if !node.config.Simulation() {
		panic(errors.New("MaximumResourcesSim request can only be used in simulation mode"))
	}
	return node.discovery.MaximumResourcesSim()
}

func (node *Node) RefreshOffersSim() {
	if !node.config.Simulation() {
		panic(errors.New("RefreshOffersSim request can only be used in simulation mode"))
	}
	node.discovery.RefreshOffersSim()
}

func (node *Node) SpreadOffersSim() {
	if !node.config.Simulation() {
		panic(errors.New("SpreadOffersSim request can only be used in simulation mode"))
	}
	node.discovery.SpreadOffersSim()
}
