package discovery

import (
	"fmt"
	"github.com/strabox/caravela/api/client"
	"github.com/strabox/caravela/node/configuration"
	"github.com/strabox/caravela/node/guid"
	"github.com/strabox/caravela/node/resources"
	"github.com/strabox/caravela/overlay"
	"time"
)

type offerSlot struct {
}

/*
Supplier handles all the logic of offering the own resources and receiving requests to deploy containers
*/
type supplier struct {
	config             *configuration.Configuration
	overlay            overlay.Overlay
	client             client.CaravelaClient
	resourcesMap       *resources.ResourcesMap // The resources<->GUID mapping
	maxResources       *resources.Resources    // The maximum resources that the Docker engine has available (FIXED value)
	availableResources *resources.Resources
	supplyingTicker    *time.Ticker
}

func newSupplier(config *configuration.Configuration, overlay overlay.Overlay, client client.CaravelaClient,
	resourcesMap *resources.ResourcesMap, maxResources resources.Resources) *supplier {
	resSupplier := &supplier{}
	resSupplier.config = config
	resSupplier.overlay = overlay
	resSupplier.client = client
	resSupplier.resourcesMap = resourcesMap
	resSupplier.maxResources = maxResources.Copy()
	resSupplier.supplyingTicker = time.NewTicker(config.SupplyingInterval)

	return resSupplier
}

func (sup *supplier) Start() {
	fmt.Println("[Supplier] Starting")

	go sup.startSupplying()
}

func (sup *supplier) startSupplying() {
	fmt.Println("[Supplier] Starting supplying ")

	for tick := range sup.supplyingTicker.C {
		destGuid, _ := sup.resourcesMap.RandomGuid(*sup.maxResources)
		remoteNode := sup.overlay.Lookup(destGuid.Bytes())
		remoteNodeGuid := guid.NewGuidBytes(destGuid.Bytes())

		sup.client.Offer(remoteNode[0].IP(), remoteNodeGuid.String(), sup.config.HostIP, "", 1, 1)
		fmt.Println("[Supplier] Resupplying...", tick, sup.resourcesMap.GetIndexableResources(*sup.maxResources).ToString())
	}
}
