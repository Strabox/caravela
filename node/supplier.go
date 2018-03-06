package node

import (
	"github.com/strabox/caravela/node/resources"
)

/*
Supplier handles all the logic of offering the own resources and receiving requests to deploy containers
*/
type Supplier struct {
	node				*Node					// Node of the supplier 
	maxResources   		*resources.Resources	// The maximum resources that the node can offer
	resourcesAvailable 	*resources.Resources 	// The current resources that the node have available
}


func NewSupplier(node *Node, maxResources resources.Resources) *Supplier {
	resSupplier := &Supplier{}
	resSupplier.node = node
	resSupplier.maxResources = &maxResources
	resSupplier.resourcesAvailable = (&maxResources).Copy()
	return resSupplier
}