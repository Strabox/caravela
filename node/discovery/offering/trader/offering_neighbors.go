package trader

import (
	"github.com/strabox/caravela/node/common"
	"sync"
)

//
type nearbyTradersOffering struct {
	successor   *common.RemoteNode
	predecessor *common.RemoteNode
	mutex       *sync.RWMutex
}

func newNeighborTradersOffering() *nearbyTradersOffering {
	return &nearbyTradersOffering{
		successor:   nil,
		predecessor: nil,
		mutex:       &sync.RWMutex{},
	}
}

func (n *nearbyTradersOffering) Successor() *common.RemoteNode {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	return n.successor
}

func (n *nearbyTradersOffering) SetSuccessor(newSuccessor *common.RemoteNode) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.successor = newSuccessor
}

func (n *nearbyTradersOffering) Predecessor() *common.RemoteNode {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	return n.predecessor
}

func (n *nearbyTradersOffering) SetPredecessor(newPredecessor *common.RemoteNode) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.predecessor = newPredecessor
}

func (n *nearbyTradersOffering) Neighbors() []*common.RemoteNode {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	res := make([]*common.RemoteNode, 2)
	res[0] = n.predecessor
	res[1] = n.successor
	return res
}
