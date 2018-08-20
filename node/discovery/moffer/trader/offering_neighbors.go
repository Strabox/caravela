package trader

import (
	"github.com/strabox/caravela/node/common"
	"sync"
)

//
type NearbyTradersOffering struct {
	successor   *common.RemoteNode
	predecessor *common.RemoteNode
	mutex       *sync.RWMutex
}

func NewNeighborTradersOffering() *NearbyTradersOffering {
	return &NearbyTradersOffering{
		successor:   nil,
		predecessor: nil,
		mutex:       &sync.RWMutex{},
	}
}

func (neigh *NearbyTradersOffering) Successor() *common.RemoteNode {
	neigh.mutex.RLock()
	defer neigh.mutex.RUnlock()

	return neigh.successor
}

func (neigh *NearbyTradersOffering) SetSuccessor(newSuccessor *common.RemoteNode) {
	neigh.mutex.Lock()
	defer neigh.mutex.Unlock()

	neigh.successor = newSuccessor
}

func (neigh *NearbyTradersOffering) Predecessor() *common.RemoteNode {
	neigh.mutex.RLock()
	defer neigh.mutex.RUnlock()

	return neigh.predecessor
}

func (neigh *NearbyTradersOffering) SetPredecessor(newPredecessor *common.RemoteNode) {
	neigh.mutex.Lock()
	defer neigh.mutex.Unlock()

	neigh.predecessor = newPredecessor
}

func (neigh *NearbyTradersOffering) Neighbors() []*common.RemoteNode {
	neigh.mutex.RLock()
	defer neigh.mutex.RUnlock()

	return []*common.RemoteNode{neigh.predecessor, neigh.successor}
}
