package node

import (
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/api"
	"github.com/strabox/caravela/api/remote"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/docker"
	nodeAPI "github.com/strabox/caravela/node/api"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/containers"
	"github.com/strabox/caravela/node/discovery"
	"github.com/strabox/caravela/node/scheduler"
	"github.com/strabox/caravela/overlay"
	"github.com/strabox/caravela/util"
)

/*
Top level structure that contains all the modules/objects that manages a CARAVELA node.
*/
type Node struct {
	config   *configuration.Configuration // System's configuration
	stopChan chan bool                    // Channel to stop the node functions

	apiServer         api.Server           // API server to to handle requests
	discovery         *discovery.Discovery // Discovery component
	scheduler         *scheduler.Scheduler // Scheduler component
	containersManager *containers.Manager  // Containers Manager component
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
	resourcesMap.Print()

	discoveryComp := discovery.NewDiscovery(config, overlay, caravelaCli, resourcesMap, *maxResources)
	containersManagerComp := containers.NewManager(config, dockerClient, discoveryComp)

	return &Node{
		config:   config,
		stopChan: make(chan bool),

		apiServer:         apiServer,
		overlay:           overlay,
		discovery:         discoveryComp,
		containersManager: containersManagerComp,
		scheduler:         scheduler.NewScheduler(config, discoveryComp, containersManagerComp, caravelaCli),
	}
}

/*
Start the node internal working
*/
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

	err = node.apiServer.Start(node.config, node) // Start CARAVELA REST API HttpServer
	if err != nil {
		return err
	}

	log.Debug(util.LogTag("Node") + "Node STARTED")

	for {
		select {
		case stop := <-node.stopChan: // Block main Goroutine until a stop message is received
			if stop {
				return nil
			}
		}
	}
	return nil
}

/*
Stop the node internal working
*/
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

func (node *Node) AddTrader(guidBytes []byte) {
	guidRes := guid.NewGUIDBytes(guidBytes)
	node.discovery.AddTrader(*guidRes)
}

/* ================================== NodeRemote ============================= */

func (node *Node) Discovery() nodeAPI.Discovery {
	return node.discovery
}

func (node *Node) Scheduler() nodeAPI.Scheduler {
	return node.scheduler
}
