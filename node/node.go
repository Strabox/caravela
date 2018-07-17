package node

import (
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/api"
	"github.com/strabox/caravela/api/remote"
	"github.com/strabox/caravela/api/rest"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/docker"
	nodeAPI "github.com/strabox/caravela/node/api"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/containers"
	"github.com/strabox/caravela/node/discovery"
	"github.com/strabox/caravela/node/scheduler"
	"github.com/strabox/caravela/node/user"
	"github.com/strabox/caravela/overlay"
	"github.com/strabox/caravela/util"
)

// Node is the top level (Entry) structure for all the functionality.
type Node struct {
	config   *configuration.Configuration // System's configuration
	stopChan chan bool                    // Channel to stop the node functions

	apiServer         api.Server           // API server component
	discovery         *discovery.Discovery // Discovery component
	scheduler         *scheduler.Scheduler // Scheduler component
	containersManager *containers.Manager  // Containers Manager component
	userManager       *user.Manager        // User Manager component
	overlay           overlay.Overlay      // Overlay component
}

func NewNode(config *configuration.Configuration, overlay overlay.Overlay, caravelaCli remote.Caravela,
	dockerClient docker.Client, apiServer api.Server) *Node {

	// Obtain the maximum resources Docker Engine has available
	maxCPUs, maxRAM := dockerClient.GetDockerCPUAndRAM()
	maxResources := resources.NewResources(maxCPUs, maxRAM)

	// Create Resources Mapping (based on the configurations)
	resourcesMap := resources.NewResourcesMap(resources.GetCpuCoresPartitions(config.CPUCoresPartitions()),
		resources.GetRamPartitions(config.RAMPartitions()))

	// Create all the internal components

	discoveryComp := discovery.NewDiscovery(config, overlay, caravelaCli, resourcesMap, *maxResources)

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

// Configuration returns the configuration that is
func (node *Node) Configuration() *configuration.Configuration {
	return node.config
}

/* ========================= SubComponent Interface ====================== */

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

	select {
	case stop := <-node.stopChan: // Block main Goroutine until a stop message is received
		if stop {
			return nil
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

/* ======================== Overlay Membership Interface =========================== */

func (node *Node) AddTrader(guidBytes []byte) {
	guidRes := guid.NewGUIDBytes(guidBytes)
	node.discovery.AddTrader(*guidRes)
}

/* ========================= Discovery Component Interface ====================== */

func (node *Node) CreateOffer(fromSupplierGUID string, fromSupplierIP string, toTraderGUID string,
	id int64, amount int, cpus int, ram int) {
	node.discovery.CreateOffer(fromSupplierGUID, fromSupplierIP, toTraderGUID, id, amount, cpus, ram)
}

func (node *Node) RefreshOffer(offerID int64, fromTraderGUID string) bool {
	return node.discovery.RefreshOffer(offerID, fromTraderGUID)
}

func (node *Node) RemoveOffer(fromSupplierIP string, fromSupplierGUID string, toTraderGUID string, offerID int64) {
	node.discovery.RemoveOffer(fromSupplierIP, fromSupplierGUID, toTraderGUID, offerID)
}

func (node *Node) GetOffers(toTraderGUID string, relay bool, fromNodeGUID string) []nodeAPI.Offer {
	return node.discovery.GetOffers(toTraderGUID, relay, fromNodeGUID)
}

func (node *Node) AdvertiseNeighborOffers(toTraderGUID string, fromTraderGUID string, traderOfferingIP string,
	traderOfferingGUID string) {
	node.discovery.AdvertiseNeighborOffers(toTraderGUID, fromTraderGUID, traderOfferingIP, traderOfferingGUID)
}

/* ========================= Scheduling Component Interface ======================== */

func (node *Node) LaunchContainers(fromBuyerIP string, offerId int64, containerImageKey string, portMappings []rest.PortMapping,
	containerArgs []string, cpus int, ram int) (string, error) {
	return node.scheduler.Launch(fromBuyerIP, offerId, containerImageKey, portMappings, containerArgs, cpus, ram)
}

/* ========================= Containers Component Interface ======================== */

func (node *Node) StopLocalContainer(containerID string) error {
	return node.containersManager.StopContainer(containerID)
}

/* ========================= User Component Interface ============================= */

func (node *Node) SubmitContainers(containerImageKey string, portMappings []rest.PortMapping, containerArgs []string,
	cpus int, ram int) error {
	return node.userManager.SubmitContainers(containerImageKey, portMappings, containerArgs, cpus, ram)
}

func (node *Node) StopContainers(containersIDs []string) error {
	return node.userManager.StopContainers(containersIDs)
}

func (node *Node) ListContainers() rest.ContainersList {
	return node.userManager.ListContainers()
}
