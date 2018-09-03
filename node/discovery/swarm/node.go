package swarm

import (
	"github.com/strabox/caravela/node/common/resources"
	"sync"
)

type node struct {
	ipAddress          string              //
	availableResources resources.Resources //
	usedResources      resources.Resources
	containersRunning  int //

	mutex sync.RWMutex //
}

func newNode(ip string, availableResources, usedResources resources.Resources) *node {
	return &node{
		ipAddress:          ip,
		availableResources: availableResources,
		usedResources:      usedResources,
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

func (n *node) setUsedResources(newUsedResources resources.Resources) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.usedResources.SetTo(newUsedResources)
}

func (n *node) ip() string {
	return n.ipAddress
}

func (n *node) totalAvailableResources() resources.Resources {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	return n.availableResources
}

func (n *node) totalUsedResources() resources.Resources {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	return n.usedResources
}

func (n *node) totalContainersRunning() int {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	return n.containersRunning
}
