package supplier

import (
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/api/remote"
	"github.com/strabox/caravela/configuration"
	nodeCommon "github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/discovery/common"
	"github.com/strabox/caravela/overlay"
	"sync"
	"time"
)

/*
Supplier handles all the logic of offering the own resources and receiving requests to deploy containers
*/
type Supplier struct {
	config             *configuration.Configuration
	overlay            overlay.Overlay
	client             remote.Caravela      // Client to collaborate with other CARAVELA's nodes
	resourcesMap       *resources.Mapping   // The resources<->GUID mapping
	maxResources       *resources.Resources // The maximum resources that the Docker engine has available (Static value)
	availableResources *resources.Resources // Available resources to offerContent

	supplyingTicker      <-chan time.Time // Timer to supply available resources
	refreshesCheckTicker <-chan time.Time // Timer to check if the offers are in alive traders

	offersID    int64                    // Monotonic counter to generate offer's local unique IDs
	offers      map[int64]*supplierOffer // Map with the current offers (that are being managed by traders)
	offersMutex *sync.Mutex              // Mutex to handle offers management
}

func NewSupplier(config *configuration.Configuration, overlay overlay.Overlay, client remote.Caravela,
	resourcesMap *resources.Mapping, maxResources resources.Resources) *Supplier {
	resSupplier := &Supplier{}
	resSupplier.config = config
	resSupplier.overlay = overlay
	resSupplier.client = client
	resSupplier.resourcesMap = resourcesMap
	resSupplier.maxResources = maxResources.Copy()
	resSupplier.availableResources = maxResources.Copy()

	resSupplier.supplyingTicker = time.NewTicker(config.SupplyingInterval()).C
	resSupplier.refreshesCheckTicker = time.NewTicker(config.RefreshesCheckInterval()).C

	resSupplier.offersID = 0
	resSupplier.offers = make(map[int64]*supplierOffer)
	resSupplier.offersMutex = &sync.Mutex{}
	return resSupplier
}

func (sup *Supplier) Start() {
	log.Debugln("[Supplier] Starting supplying resources...")
	go sup.startSupplying()
}

func (sup *Supplier) startSupplying() {
	for {
		select {
		case <-sup.supplyingTicker: // Offer the available resources into a random trader (responsible for them)
			sup.offersMutex.Lock()

			if !sup.availableResources.IsZero() {
				log.Debugln("[Supplier] Resupplying...", sup.resourcesMap.GetResourcesIndexes(*sup.maxResources).String())

				destinationGUID, _ := sup.resourcesMap.RandomGuid(*sup.availableResources)
				remoteNode := sup.overlay.Lookup(destinationGUID.Bytes())
				remoteNodeGUID := guid.NewGuidBytes(remoteNode[0].GUID())

				/*
					handledResources, _ := sup.resourcesMap.ResourcesByGuid(*remoteNodeGUID)
					for !sup.availableResources.Contains(*handledResources) {

					}
				*/

				go func() {
					sup.offersMutex.Lock()
					defer sup.offersMutex.Unlock()

					err := sup.client.CreateOffer(sup.config.HostIP(), "", remoteNode[0].IP(), remoteNodeGUID.String(),
						sup.offersID, 1, sup.availableResources.CPU(), sup.availableResources.RAM())

					if err == nil {
						sup.offers[sup.offersID] = newSupplierOffer(common.NewOffer(common.OfferID(sup.offersID),
							1, *sup.availableResources.Copy()), *remoteNodeGUID)
						sup.availableResources.SetZero()
						sup.offersID++
					}
				}()
			}

			sup.offersMutex.Unlock()
		case <-sup.refreshesCheckTicker: // Check if the offers are being refreshed by the respective trader
			sup.offersMutex.Lock()

			log.Debugln("[Supplier] Checking refreshes...")

			for offerKey, offer := range sup.offers {

				offer.VerifyRefreshMiss(sup.config.RefreshMissedTimeout())

				if offer.RefreshesMissed() >= sup.config.MaxRefreshesMissed() {
					sup.availableResources.Add(*offer.Resources())
					delete(sup.offers, offerKey)
				}
			}

			sup.offersMutex.Unlock()
		}
	}
}

func (sup *Supplier) FindNodes(resources resources.Resources) []*nodeCommon.RemoteNode {
	sup.offersMutex.Lock()
	defer sup.offersMutex.Unlock()

	var resultNodes []*nodeCommon.RemoteNode = nil

	destinationGUID, _ := sup.resourcesMap.RandomGuid(resources)
	overlayNodes := sup.overlay.Lookup(destinationGUID.Bytes())

	if overlayNodes != nil {
		resultNodes = make([]*nodeCommon.RemoteNode, len(overlayNodes))
		for i, overlayNode := range overlayNodes {
			nodeGuid := guid.NewGuidBytes(overlayNode.GUID())
			resultNodes[i] = nodeCommon.NewRemoteNode(overlayNode.IP(), *nodeGuid)
		}
	}

	return resultNodes
}

func (sup *Supplier) RefreshOffer(offerID int64, fromTraderGUID string) bool {
	sup.offersMutex.Lock()
	defer sup.offersMutex.Unlock()
	log.Debugf("[Supplier] Refreshing Offer: %d from: %s", offerID, fromTraderGUID)

	offer, exist := sup.offers[offerID]

	if exist {
		if offer.IsResponsibleTrader(*guid.NewGuidString(fromTraderGUID)) {
			offer.Refresh()
			log.Debugf("[Supplier] Offer: %d refresh SUCCESS", offerID)
			return true
		} else {
			log.Debugf("[Supplier] Offer: %d refresh FAILED (Wrong trader)", offerID)
			return false // TODO: Return an error for fake traders trying to refresh the offer?
		}
	} else {
		log.Debugf("[Supplier] Offer: %d refresh FAILED (Offer does not exist)", offerID)
		return false // TODO: Return an error because this offerContent does not exist and the trader can remove it?
	}
}
