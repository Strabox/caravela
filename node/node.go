//Node represents a chord node and is the manager of the node
package node

import (
	"github.com/strabox/caravela/api/client"
	"github.com/strabox/caravela/configuration"
	nodeAPI "github.com/strabox/caravela/node/api"
	"github.com/strabox/caravela/node/discovery"
	"github.com/strabox/caravela/node/guid"
	"github.com/strabox/caravela/node/resources"
	"github.com/strabox/caravela/node/scheduler"
	"github.com/strabox/caravela/overlay"
)

type Node struct {
	discovery *discovery.Discovery
	scheduler *scheduler.Scheduler
}

func NewNode(config *configuration.Configuration, overlay overlay.Overlay, client client.Caravela,
	maxResources resources.Resources) *Node {

	// Resources Mapping creation based on the configurations
	resourcesMap := resources.NewResourcesMap(config.CpuPartitions(), config.RamPartitions())
	resourcesMap.Print()

	res := &Node{}
	res.discovery = discovery.NewDiscovery(config, overlay, client, resourcesMap, maxResources)
	res.scheduler = scheduler.NewScheduler()
	return res
}

func (node *Node) Start() {
	node.discovery.Start()
}

func (node *Node) AddTrader(guidBytes []byte) {
	guidRes := guid.NewGuidBytes(guidBytes)
	node.discovery.AddTrader(*guidRes)
}

/* ================================== NodeRemote ============================= */

func (node *Node) Discovery() nodeAPI.Discovery {
	return node.discovery
}

func (node *Node) Scheduler() nodeAPI.Scheduler {
	return node.scheduler
}
