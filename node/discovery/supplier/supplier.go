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
	"github.com/strabox/caravela/util"
	"sync"
	"time"
)

/*
Supplier handles all the logic of managing the node own resources, advertising them into the system.
*/
type Supplier struct {
	config  *configuration.Configuration // Configurations of the system
	overlay overlay.Overlay              // Node overlay to efficient route/search of messages
	client  remote.Caravela              // Client to collaborate with other CARAVELA's nodes

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

/*
Starts the supplier.
*/
func (sup *Supplier) Start() {
	log.Debug(util.LogTag("[Supplier]") + "STARTED...")
	go sup.startSupplying()
}

/*
Controls the time dependant actions like supplying the resources.
*/
func (sup *Supplier) startSupplying() {
	for {
		select {
		/*
			- Offer the available resources into a random trader (responsible for them).
			- This actions can be customized with a different strategy to split the available resources
			into different resource offers.
		*/
		case <-sup.supplyingTicker:
			sup.offersMutex.Lock()

			if !sup.availableResources.IsZero() {
				log.Debugf(util.LogTag("[Supplier]")+"Supplying: %s",
					sup.resourcesMap.GetFittestResources(*sup.maxResources).String())

				destinationGUID, _ := sup.resourcesMap.RandomGUID(*sup.availableResources)
				remoteNode := sup.overlay.Lookup(destinationGUID.Bytes())
				remoteNodeGUID := guid.NewGuidBytes(remoteNode[0].GUID())

				err := sup.client.CreateOffer(sup.config.HostIP(), "", remoteNode[0].IP(),
					remoteNodeGUID.String(), sup.offersID, 1, sup.availableResources.CPU(),
					sup.availableResources.RAM())

				if err == nil {
					sup.offers[sup.offersID] = newSupplierOffer(common.OfferID(sup.offersID),
						1, *sup.availableResources, *remoteNodeGUID)
					sup.availableResources.SetZero()
					sup.offersID++
				}
			}

			sup.offersMutex.Unlock()
			// Check if the offers are being refreshed by the respective trader
		case <-sup.refreshesCheckTicker:
			sup.offersMutex.Lock()
			log.Debug(util.LogTag("[Supplier]") + "Checking refreshes...")

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

/*
Finds a list of CARAVELA's nodes (in this case Traders) that manage the given resources combinations.
*/
func (sup *Supplier) FindNodes(resources resources.Resources) []*nodeCommon.RemoteNode {
	var resultNodes []*nodeCommon.RemoteNode = nil

	destinationGUID, _ := sup.resourcesMap.RandomGUID(resources)
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

/*
Refreshes an offer.
*/
func (sup *Supplier) RefreshOffer(offerID int64, fromTraderGUID string) bool {
	sup.offersMutex.Lock()
	defer sup.offersMutex.Unlock()
	log.Debugf(util.LogTag("[Supplier]")+"Refreshing Offer: %d from: %s", offerID, fromTraderGUID)

	offer, exist := sup.offers[offerID]

	if exist {
		if offer.IsResponsibleTrader(*guid.NewGuidString(fromTraderGUID)) {
			offer.Refresh()
			log.Debugf(util.LogTag("[Supplier]")+"Offer: %d refresh SUCCESS", offerID)
			return true
		} else {
			log.Debugf(util.LogTag("[Supplier]")+"Offer: %d refresh FAILED (Wrong trader)", offerID)
			return false // TODO: Return an error for fake traders trying to refresh the offer?
		}
	} else {
		log.Debugf(util.LogTag("[Supplier]")+"Offer: %d refresh FAILED (Offer does not exist)", offerID)
		return false // TODO: Return an error because this offerContent does not exist and the trader can remove it?
	}
}

func (sup *Supplier) ObtainOffer(offerID int64) *resources.Resources {
	sup.offersMutex.Lock()
	defer sup.offersMutex.Unlock()

	supOffer := sup.offers[offerID]
	if supOffer == nil {
		return nil
	} else {
		res := sup.offers[offerID].Resources().Copy()
		delete(sup.offers, offerID)
		return res
	}
}

func (sup *Supplier) ReturnOffer(returnedResources resources.Resources) {
	sup.offersMutex.Lock()
	defer sup.offersMutex.Unlock()

	sup.availableResources.Add(returnedResources)
}
