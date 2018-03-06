//Node represents a chord node and is the manager of the node
package node

import (
	"github.com/strabox/caravela/overlay"
)

type Node struct {
	guid   *Supplier
	overlay      overlay.Overlay
	resourcesMap *ResourcesMap
	supplier     *Supplier
	traders      []*Trader
}

func NewNode(o overlay.Overlay, rm *ResourcesMap, sup *Supplier) *Node {
	res := &Node{}
	res.overlay = o
	res.resourcesMap = rm
	res.supplier = sup
	res.traders = nil
	return res
}

func NewNodeTraders(o overlay.Overlay, rm *ResourcesMap, sup *Supplier, traders []*Trader) *Node {
	res := &Node{}
	res.overlay = o
	res.resourcesMap = rm
	res.supplier = sup
	res.traders = traders
	return res
}

func (node *Node) ResourcesMap() *ResourcesMap {
	return node.resourcesMap
	overlay      overlay.Overlay
	resourcesMap *ResourcesMap
	supplier     *Supplier
	traders      []*Trader
}

func NewNode(o overlay.Overlay, rm *ResourcesMap, sup *Supplier) *Node {
	res := &Node{}
	res.overlay = o
	res.resourcesMap = rm
	res.supplier = sup
	res.traders = nil
	return res
}

func NewNodeTraders(o overlay.Overlay, rm *ResourcesMap, sup *Supplier, traders []*Trader) *Node {
	res := &Node{}
	res.overlay = o
	res.resourcesMap = rm
	res.supplier = sup
	res.traders = traders
	return res
}

func (node *Node) ResourcesMap() *ResourcesMap {
	return node.resourcesMap
}
