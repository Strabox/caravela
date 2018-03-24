package supplier

import (
	"github.com/strabox/caravela/api/client"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/guid"
	"github.com/strabox/caravela/node/resources"
	"github.com/strabox/caravela/overlay"
	"log"
	"sync"
	"time"
)

type offerSupplier struct {
	offerID   int
	amount    int
	resources *resources.Resources

	missingRefreshes int
}

/*
Supplier handles all the logic of offering the own resources and receiving requests to deploy containers
*/
type Supplier struct {
	config             *configuration.Configuration
	overlay            overlay.Overlay
	client             client.Caravela
	resourcesMap       *resources.Mapping   // The resources<->GUID mapping
	maxResources       *resources.Resources // The maximum resources that the Docker engine has available (Static value)
	availableResources *resources.Resources // Available resources to offer

	supplyingTicker *time.Ticker

	offersID    int
	offers      map[int]*offerSupplier
	offersMutex *sync.Mutex
}

func NewSupplier(config *configuration.Configuration, overlay overlay.Overlay, client client.Caravela,
	resourcesMap *resources.Mapping, maxResources resources.Resources) *Supplier {
	resSupplier := &Supplier{}
	resSupplier.config = config
	resSupplier.overlay = overlay
	resSupplier.client = client
	resSupplier.resourcesMap = resourcesMap
	resSupplier.maxResources = maxResources.Copy()
	resSupplier.availableResources = maxResources.Copy()

	resSupplier.supplyingTicker = time.NewTicker(config.SupplyingInterval())

	resSupplier.offersID = 0
	resSupplier.offers = make(map[int]*offerSupplier)
	resSupplier.offersMutex = &sync.Mutex{}
	return resSupplier
}

func (sup *Supplier) Start() {
	log.Println("[Supplier] Starting supplying resource's offers...")

	go sup.startSupplying()
}

func (sup *Supplier) startSupplying() {
	for range sup.supplyingTicker.C {
		sup.offersMutex.Lock()

		if !sup.availableResources.IsZero() {
			destinationGUID, _ := sup.resourcesMap.RandomGuid(*sup.availableResources)
			remoteNode := sup.overlay.Lookup(destinationGUID.Bytes())
			remoteNodeGuid := guid.NewGuidBytes(remoteNode[0].Guid())

			err := sup.client.CreateOffer(sup.config.HostIP(), "", remoteNode[0].IP(), remoteNodeGuid.String(),
				sup.offersID, 1, sup.availableResources.CPU(), sup.availableResources.RAM())

			if err == nil {
				sup.offers[sup.offersID] = &offerSupplier{sup.offersID, 1, sup.availableResources.Copy(),
					0}
				sup.availableResources.SetZero()
				sup.offersID++
			}
			log.Println("[Supplier] Resupplying...", sup.resourcesMap.GetIndexableResources(*sup.maxResources).String())
		}

		sup.offersMutex.Unlock()
	}
}

func (sup *Supplier) RefreshOffer(id int, fromTraderGUID string) bool {
	return true
}
