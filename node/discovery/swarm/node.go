package swarm

import (
	"github.com/strabox/caravela/node/common/resources"
	"sync"
)

// node represents a node in the cluster.
// It maintains the most updated information about the node.
type node struct {
	ipAddress         string              // Node's IP address.
	freeResources     resources.Resources // Node's free resources.
	usedResources     resources.Resources // Node's used resources.
	containersRunning int                 // Node's current number of containers running.
	mutex             sync.RWMutex        // Node's mutex.
}

// newNode creates a new node based on the current node information.
func newNode(ip string, freeResources, usedResources resources.Resources) *node {
	return &node{
		ipAddress:         ip,
		freeResources:     freeResources,
		usedResources:     usedResources,
		containersRunning: 0,
		mutex:             sync.RWMutex{},
	}
}

// setContainerRunning sets the number of container running in the node.
func (n *node) setContainerRunning(containersRunning int) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.containersRunning = containersRunning
}

// setContainerRunning sets the amount of free resources in the node.
func (n *node) setFreeResources(newFreeResources resources.Resources) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.freeResources.SetTo(newFreeResources)
}

// setUsedResources sets the amount of used resources in the node.
func (n *node) setUsedResources(newUsedResources resources.Resources) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.usedResources.SetTo(newUsedResources)
}

// ip returns the node's ip address.
func (n *node) ip() string {
	return n.ipAddress
}

// totalFreeResources returns the current amount of free resources in the node.
func (n *node) totalFreeResources() resources.Resources {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	return n.freeResources
}

// totalUsedResources returns the current amount of used resources in the node.
func (n *node) totalUsedResources() resources.Resources {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	return n.usedResources
}

// totalContainersRunning returns the current amount of containers running in the node.
func (n *node) totalContainersRunning() int {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	return n.containersRunning
}
