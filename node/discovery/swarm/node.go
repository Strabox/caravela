package swarm

import (
	"github.com/strabox/caravela/node/common/resources"
	"sync"
)

// clusterNode represents a clusterNode in the cluster.
// It maintains the most updated information about the clusterNode.
type clusterNode struct {
	ipAddress         string              // Node's IP address.
	freeResources     resources.Resources // Node's free resources.
	usedResources     resources.Resources // Node's used resources.
	containersRunning int                 // Node's current number of containers running.
	mutex             sync.RWMutex        // Node's mutex.
}

// newClusterNode creates a new clusterNode based on the current clusterNode information.
func newClusterNode(ip string, freeResources, usedResources resources.Resources) *clusterNode {
	return &clusterNode{
		ipAddress:         ip,
		freeResources:     freeResources,
		usedResources:     usedResources,
		containersRunning: 0,
		mutex:             sync.RWMutex{},
	}
}

// setContainerRunning sets the number of container running in the clusterNode.
func (n *clusterNode) setContainerRunning(containersRunning int) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.containersRunning = containersRunning
}

// setContainerRunning sets the amount of free resources in the clusterNode.
func (n *clusterNode) setFreeResources(newFreeResources resources.Resources) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.freeResources.SetTo(newFreeResources)
}

// setUsedResources sets the amount of used resources in the clusterNode.
func (n *clusterNode) setUsedResources(newUsedResources resources.Resources) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.usedResources.SetTo(newUsedResources)
}

// ip returns the clusterNode's ip address.
func (n *clusterNode) ip() string {
	return n.ipAddress
}

// totalFreeResources returns the current amount of free resources in the clusterNode.
func (n *clusterNode) totalFreeResources() resources.Resources {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	return n.freeResources
}

// totalUsedResources returns the current amount of used resources in the clusterNode.
func (n *clusterNode) totalUsedResources() resources.Resources {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	return n.usedResources
}

// totalContainersRunning returns the current amount of containers running in the clusterNode.
func (n *clusterNode) totalContainersRunning() int {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	return n.containersRunning
}
