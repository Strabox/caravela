package supplier

import (
	"github.com/strabox/caravela/node/resources"
	"github.com/strabox/caravela/node/guid"
	"github.com/strabox/caravela/overlay"
	"github.com/strabox/caravela/node/discovery"
	"github.com/strabox/caravela/node/configuration"
	"github.com/strabox/caravela/api/client"
	"fmt"
	"time"
)

/*
Discovery has all the operations that the resources discovery/supplier module should have
*/
type Discovery interface {
	Discover(r resources.Resources) []overlay.RemoteNode
	AcceptTrader(offer discovery.OfferID, guid *guid.Guid, ip string)
}

/*
Supplier handles all the logic of offering the own resources and receiving requests to deploy containers
*/
type Supplier struct {
	config				*configuration.Configuration
	overlay      		overlay.Overlay
	client				client.CaravelaClient
	resourcesMap 		*resources.ResourcesMap
	maxResources   		*resources.Resources			// The maximum resources that the Docker engine has available
	resourcesAvailable 	*resources.Resources 			// The non reserved resources
	supplyingTicker		*time.Ticker
}



func NewSupplier(config *configuration.Configuration,overlay overlay.Overlay, client client.CaravelaClient, 
		resourcesMap *resources.ResourcesMap, maxResources resources.Resources) *Supplier {
	resSupplier := &Supplier{}
	resSupplier.config = config
	resSupplier.overlay = overlay
	resSupplier.client = client
	resSupplier.resourcesMap = resourcesMap
	resSupplier.maxResources = maxResources.Copy()
	resSupplier.resourcesAvailable = maxResources.Copy()
	resSupplier.supplyingTicker = time.NewTicker(config.SupplyingInterval)
	
	go resSupplier.StartSupplying()
	
	return resSupplier
}

func (sup* Supplier) StartSupplying() {
	fmt.Println("[Supplier] Starting supplying ")
	
	for tick := range sup.supplyingTicker.C {
		destGuid, _ := sup.resourcesMap.RandomGuid(*sup.resourcesAvailable)
		remoteNode := sup.overlay.Lookup(*destGuid)
		
		sup.client.Offer(remoteNode[0].IP(), remoteNode[0].Guid().String(), sup.config.HostIP, 1, 1)
        fmt.Println("[Supplier] Resupplying...", tick, sup.resourcesMap.GetIndexableResources(*sup.resourcesAvailable).ToString())
    }
}