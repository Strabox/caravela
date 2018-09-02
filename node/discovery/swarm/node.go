package swarm

import (
	"github.com/strabox/caravela/node/common/resources"
	"sync"
)

type node struct {
	ipAddress          string              //
	availableResources resources.Resources //
	containersRunning  int                 //

	mutex sync.RWMutex //
}

func newNode(ip string, availableResources resources.Resources) *node {
	return &node{
		ipAddress:          ip,
		availableResources: availableResources,
		containersRunning:  0,

		mutex: sync.RWMutex{},
	}
}

func (n *node) setContainerRunning(containersRunning int) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.containersRunning = containersRunning
}

func (n *node) setAvailableResources(newAvailableResources resources.Resources) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.availableResources.SetTo(newAvailableResources)
}

func (n *node) ip() string {
	return n.ipAddress
}

func (n *node) totalAvailableResources() resources.Resources {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	return n.availableResources
}

func (n *node) totalContainersRunning() int {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	return n.containersRunning
}
