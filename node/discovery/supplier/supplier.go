package supplier

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/api/remote"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/api"
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
	nodeCommon.SystemSubComponent // Base component

	config  *configuration.Configuration // Configurations of the system
	overlay overlay.Overlay              // Node overlay to efficient route messages to specific nodes.
	client  remote.Caravela              // Client to collaborate with other CARAVELA's nodes

	resourcesMap       *resources.Mapping                // The resources<->GUID mapping
	maxResources       *resources.Resources              // The maximum resources that the Docker engine has available (Static value)
	availableResources *resources.Resources              // CURRENT Available resources to offer
	offersID           common.OfferID                    // Monotonic counter to generate offer's local unique IDs
	activeOffers       map[common.OfferID]*supplierOffer // Map with the current activeOffers (that are being managed by traders)
	offersMutex        *sync.Mutex                       // Mutex to handle active offers management

	quitChan             chan bool        // Channel to alert that the node is stopping
	supplyingTicker      <-chan time.Time // Timer to supply available resources
	refreshesCheckTicker <-chan time.Time // Timer to check if the active offers are in alive traders
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
	resSupplier.offersID = 0
	resSupplier.activeOffers = make(map[common.OfferID]*supplierOffer)
	resSupplier.offersMutex = &sync.Mutex{}

	resSupplier.quitChan = make(chan bool)
	resSupplier.supplyingTicker = time.NewTicker(config.SupplyingInterval()).C
	resSupplier.refreshesCheckTicker = time.NewTicker(config.RefreshesCheckInterval()).C
	return resSupplier
}

/*
Controls the time dependant actions like supplying the resources.
*/
func (sup *Supplier) startSupplying() {
	for {
		select {
		case <-sup.supplyingTicker: // Offer the available resources into a random trader (responsible for them).
			go func() {
				// TODO: OPTIONAL ENHANCE the parallelism in this call
				sup.offersMutex.Lock()
				defer sup.offersMutex.Unlock()

				if sup.availableResources.Available() {
					/*
						What?: Remove all active offers from the traders in order to gather all available resources.
						Goal: This is used to try offer the maximum amount of resources the node has available between
							  the Available (offered) and the Available (but not offered).
					*/
					for offerID, offer := range sup.activeOffers {
						go func(offerID int64, offer *supplierOffer) {
							sup.client.RemoveOffer(sup.config.HostIP(), "", offer.ResponsibleTraderIP(),
								offer.ResponsibleTraderGUID().String(), offerID)
						}(int64(offerID), offer) // Send remove offer message in background

						delete(sup.activeOffers, offerID)
						sup.availableResources.Add(*offer.Resources())
					}

					var err error
					var overlayNodes []*overlay.Node = nil
					destinationGUID, _ := sup.resourcesMap.RandGUID(*sup.availableResources)
					overlayNodes, _ = sup.overlay.Lookup(destinationGUID.Bytes())
					overlayNodes = sup.removeNonTargetNodes(overlayNodes, *destinationGUID)

					// .. try search nodes in the beginning of the original target resource range region
					if len(overlayNodes) == 0 {
						destinationGUID := sup.resourcesMap.FirstGUID(*sup.availableResources)
						overlayNodes, _ = sup.overlay.Lookup(destinationGUID.Bytes())
						overlayNodes = sup.removeNonTargetNodes(overlayNodes, *destinationGUID)
					}

					// ... try search for random nodes that handle less powerful resource combinations
					for len(overlayNodes) == 0 {
						destinationGUID, err = sup.resourcesMap.LowerRandGUID(*destinationGUID, *sup.availableResources)
						if err != nil {
							log.Errorf(util.LogTag("Supplier")+"NO NODES to handle resources offer: %s, error: %s",
								sup.availableResources.String(), err)
							return // Wait fot the next tick to try supply resources
						}
						overlayNodes, _ = sup.overlay.Lookup(destinationGUID.Bytes())
						overlayNodes = sup.removeNonTargetNodes(overlayNodes, *destinationGUID)
					}

					// Chose the first node returned by the overlay API
					chosenNode := overlayNodes[0]
					chosenNodeGUID := guid.NewGUIDBytes(chosenNode.GUID())

					err = sup.client.CreateOffer(sup.config.HostIP(), "", chosenNode.IP(),
						chosenNodeGUID.String(), int64(sup.offersID), 1, sup.availableResources.CPUs(),
						sup.availableResources.RAM())

					if err == nil {
						sup.activeOffers[sup.offersID] = newSupplierOffer(common.OfferID(sup.offersID),
							1, *sup.availableResources, chosenNode.IP(), *chosenNodeGUID)
						sup.availableResources.SetZero()
						sup.offersID++
					}
				}
			}()
		case <-sup.refreshesCheckTicker: // Check if the activeOffers are being refreshed by the respective trader
			go func() {
				sup.offersMutex.Lock()
				defer sup.offersMutex.Unlock()

				for offerKey, offer := range sup.activeOffers {
					offer.VerifyRefreshes(sup.config.RefreshMissedTimeout())

					if offer.RefreshesMissed() >= sup.config.MaxRefreshesMissed() {
						log.Debugf(util.LogTag("Supplier")+"OFFER DOWN, ID: %d, ResponsibleTrader: %s",
							offer.ID(), offer.ResponsibleTraderIP())

						sup.availableResources.Add(*offer.Resources())
						delete(sup.activeOffers, offerKey)
					}
				}
			}()
		case res := <-sup.quitChan: // Stopping the supplier
			if res {
				log.Infof(util.LogTag("Supplier") + "STOPPED")
				return
			}
		}
	}
}

/*
Find a list active Offers that best suit the target resources given.
*/
func (sup *Supplier) FindOffers(targetResources resources.Resources) []api.Offer {
	if !sup.isWorking() {
		panic(fmt.Errorf("can't find offers, supplier not working"))
	}

	var destinationGUID *guid.GUID = nil
	findPhase := 0
	for {
		var err error = nil

		if findPhase == 0 { // Random trader inside resources zone
			destinationGUID, _ = sup.resourcesMap.RandGUID(targetResources)
		} else { // Random trader in higher resources zone
			destinationGUID, err = sup.resourcesMap.HigherRandGUID(*destinationGUID, targetResources)
			if err != nil {
				return make([]api.Offer, 0)
			} // No more resource partitions to search
		}

		res, _ := sup.resourcesMap.ResourcesByGUID(*destinationGUID)
		log.Debugf("DestinationGUIDRes: %s", res.String())

		overlayNodes, _ := sup.overlay.Lookup(destinationGUID.Bytes())
		overlayNodes = sup.removeNonTargetNodes(overlayNodes, *destinationGUID)

		for _, node := range overlayNodes {
			offers, err := sup.client.GetOffers(node.IP(), guid.NewGUIDBytes(node.GUID()).String(), true, "")
			if (err == nil) && (len(offers) != 0) {
				return offers
			}
		}
		findPhase++
	}
}

/*
Tries refresh an offer. Called when a refresh message was received.
*/
func (sup *Supplier) RefreshOffer(offerID int64, fromTraderGUID string) bool {
	if !sup.isWorking() {
		panic(fmt.Errorf("can't refresh offer, supplier not working"))
	}

	sup.offersMutex.Lock()
	defer sup.offersMutex.Unlock()

	offer, exist := sup.activeOffers[common.OfferID(offerID)]

	if !exist {
		log.Debugf(util.LogTag("Supplier")+"ID: %d refresh FAILED (Offer does not exist)", offerID)
		return false
	}

	if offer.IsResponsibleTrader(*guid.NewGUIDString(fromTraderGUID)) {
		offer.Refresh()
		log.Debugf(util.LogTag("Supplier")+"ID: %d refresh SUCCESS", offerID)
		return true
	} else {
		log.Debugf(util.LogTag("Supplier")+"ID: %d refresh FAILED (wrong trader)", offerID)
		return false
	}
}

/*
Tries to obtain a subset of the resources represented by the given offer in order to deploy  a container.
It updates the respective trader that manages the offer.
*/
func (sup *Supplier) ObtainResources(offerID int64, resourcesNecessary resources.Resources) bool {
	if !sup.isWorking() {
		panic(fmt.Errorf("can't obtain resources, supplier not working"))
	}

	sup.offersMutex.Lock()
	defer sup.offersMutex.Unlock()

	supOffer, exist := sup.activeOffers[common.OfferID(offerID)]

	// Offer does not exist in the supplier OR asking more resources than the offer has available
	if !exist || !supOffer.Resources().Contains(resourcesNecessary) {
		return false
	} else {
		remainingResources := supOffer.Resources().Copy()
		remainingResources.Sub(resourcesNecessary)

		sup.availableResources.Add(*remainingResources)

		delete(sup.activeOffers, common.OfferID(offerID))
		go func() {
			sup.client.RemoveOffer(sup.config.HostIP(), "", supOffer.ResponsibleTraderIP(),
				supOffer.ResponsibleTraderGUID().String(), int64(supOffer.ID()))
		}() // Send remove offer message in background

		return true
	}
}

/*
Release resources of an used offer into the supplier again in order to offer them again into the system.
*/
func (sup *Supplier) ReturnResources(releasedResources resources.Resources) {
	if !sup.isWorking() {
		panic(fmt.Errorf("can't return resources, supplier not working"))
	}

	sup.offersMutex.Lock()
	defer sup.offersMutex.Unlock()

	sup.availableResources.Add(releasedResources)
}

/*
Remove nodes that do not belong to that target GUID partition. (Probably because we were target a frontier node)
*/
func (sup *Supplier) removeNonTargetNodes(remoteNodes []*overlay.Node, targetGuid guid.GUID) []*overlay.Node {
	resultNodes := make([]*overlay.Node, 0)
	targetGuidResources, _ := sup.resourcesMap.ResourcesByGUID(targetGuid)
	for _, remoteNode := range remoteNodes {
		remoteNodeResources, _ := sup.resourcesMap.ResourcesByGUID(*guid.NewGUIDBytes(remoteNode.GUID()))
		if targetGuidResources.Equals(*remoteNodeResources) {
			resultNodes = append(resultNodes, remoteNode)
		}
	}
	return resultNodes
}

/*
===============================================================================
							SubComponent Interface
===============================================================================
*/

func (sup *Supplier) Start() {
	sup.Started(func() {
		go sup.startSupplying()
	})
}

func (sup *Supplier) Stop() {
	sup.Started(func() {
		sup.quitChan <- true
	})
}

func (sup *Supplier) isWorking() bool {
	return sup.Working()
}
