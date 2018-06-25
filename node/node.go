package node

import (
	"context"
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
	"github.com/strabox/caravela/overlay/chord"
	"github.com/strabox/caravela/util"
	"net/http"
)

/*
Top level structure that contains all the modules/objects that manages a CARAVELA node.
*/
type Node struct {
	config    *configuration.Configuration // System's configuration
	apiServer *http.Server                 // HTTP server to to handle API requests

	discovery         *discovery.Discovery // Discovery component
	scheduler         *scheduler.Scheduler // Scheduler component
	containersManager *containers.Manager  // Containers Manager component
	overlay           overlay.Overlay      // Overlay component

	stopChan chan bool // Channel to stop the node functions
}

func NewNode(config *configuration.Configuration) *Node {
	res := &Node{}
	res.config = config
	res.stopChan = make(chan bool)

	// Global GUID size initialization
	guid.InitializeGUID(config.ChordHashSizeBits())

	// Create Overlay Component (Chord overlay initial)
	res.overlay = chord.NewChordOverlay(guid.SizeBytes(), config.HostIP(), config.OverlayPort(),
		config.ChordVirtualNodes(), config.ChordNumSuccessors(), config.ChordTimeout())

	// Create CARAVELA's Remote Client
	caravelaCli := remote.NewHttpClient(config)

	// Create Resources Mapping (based on the configurations)
	resourcesMap := resources.NewResourcesMap(resources.GetCpuCoresPartitions(config.CPUCoresPartitions()),
		resources.GetRamPartitions(config.RAMPartitions()))
	resourcesMap.Print()

	// Create Docker client and obtain the maximum resources Docker Engine has available
	dockerClient := docker.NewDockerClient(res.config)
	maxCPUs, maxRAM := dockerClient.GetDockerCPUAndRAM()
	maxResources := resources.NewResources(maxCPUs, maxRAM)

	// Create the Resources Discovery component
	res.discovery = discovery.NewDiscovery(config, res.overlay, caravelaCli, resourcesMap, *maxResources)

	// Create the Containers Manager component
	res.containersManager = containers.NewManager(config, dockerClient, res.discovery)

	// Create the Scheduler component
	res.scheduler = scheduler.NewScheduler(config, res.discovery, res.containersManager, caravelaCli)
	return res
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

	node.apiServer, err = api.Start(node.config, node) // Start CARAVELA REST API Server
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
	go node.apiServer.Shutdown(context.Background())
	log.Debug(util.LogTag("Node") + "-> API SERVER STOPPED")
	node.scheduler.Stop()
	log.Debug(util.LogTag("Node") + "-> SCHEDULER STOPPED")
	node.containersManager.Stop()
	log.Debug(util.LogTag("Node") + "-> CONTAINERS MANAGER STOPPED")
	node.discovery.Stop()
	log.Debug(util.LogTag("Node") + "-> DISCOVERY STOPPED")
	node.overlay.Leave()
	log.Debug(util.LogTag("Node") + "-> OVERLAY STOPPED")
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
