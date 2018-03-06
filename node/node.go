//Node represents a chord node and is the manager of the node
package node

import (
	"github.com/strabox/caravela/overlay"
	"github.com/strabox/caravela/node/resources"
)

type Node struct {
	overlay      overlay.Overlay
	resourcesMap *resources.ResourcesMap
	supplier     *Supplier
	traders      []*Trader
}

func NewNode(o overlay.Overlay, rm *resources.ResourcesMap, maxNumTraders int) *Node {
	res := &Node{}
	res.overlay = o
	res.resourcesMap = rm
	res.traders = make([]*Trader, maxNumTraders)
	for index, _ := range res.traders{
		res.traders[index] = nil
	}
	return res
}

func (node *Node) ResourcesMap() *resources.ResourcesMap {
	return node.resourcesMap
}

func (node *Node) Overlay() overlay.Overlay {
	return node.overlay
}

func (node *Node) SetSupplier(sup *Supplier)  {
	node.supplier = sup
}

func (node *Node) AddTrader(trader *Trader)  {
	for index, value := range node.traders {
		if value == nil {
			node.traders[index] = trader
		}
	}
}
