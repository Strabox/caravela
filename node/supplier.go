package node

import (
	"github.com/strabox/caravela/node/resources"
	"fmt"
	"time"
)

/*
Supplier handles all the logic of offering the own resources and receiving requests to deploy containers
*/
type Supplier struct {
	node				*Node					// Node of the supplier 
	maxResources   		*resources.Resources	// The maximum resources that the Docker engine has available
	resourcesAvailable 	*resources.Resources 	// The non reserved resources
	supplyingTicker		*time.Ticker
}


func NewSupplier(node *Node, maxResources resources.Resources) *Supplier {
	resSupplier := &Supplier{}
	resSupplier.node = node
	resSupplier.maxResources = maxResources.Copy()
	resSupplier.resourcesAvailable = maxResources.Copy()
	resSupplier.supplyingTicker = time.NewTicker(5 * time.Second)
	go resSupplier.StartSupplying()
	return resSupplier
}

func (s* Supplier) StartSupplying() {
	fmt.Println("[Supplier] Starting supplying ")
	for tick := range s.supplyingTicker.C {
        fmt.Println("[Supplier] Resupplying...", tick, s.node.ResourcesMap().GetIndexableResources(*s.resourcesAvailable).ToString())
    }
}