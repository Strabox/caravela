package supplier

import (
	"fmt"
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
	nodeCommon.SystemSubComponent // Base component

	config  *configuration.Configuration // Configurations of the system
	overlay overlay.Overlay              // Node overlay to efficient route messages to specific nodes.
	client  remote.Caravela              // Client to collaborate with other CARAVELA's nodes

	resourcesMap       *resources.Mapping                // The resources<->GUID mapping
	maxResources       *resources.Resources              // The maximum resources that the Docker engine has available (Static value)
	availableResources *resources.Resources              // Available resources to offerContent
	offersID           common.OfferID                    // Monotonic counter to generate offer's local unique IDs
	activeOffers       map[common.OfferID]*supplierOffer // Map with the current activeOffers (that are being managed by traders)
	offersMutex        *sync.Mutex                       // Mutex to handle activeOffers management

	quitChan             chan bool        // Channel to alert that the node is stopping
	supplyingTicker      <-chan time.Time // Timer to supply available resources
	refreshesCheckTicker <-chan time.Time // Timer to check if the activeOffers are in alive traders
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

	resSupplier.quitChan = make(chan bool)
	resSupplier.supplyingTicker = time.NewTicker(config.SupplyingInterval()).C
	resSupplier.refreshesCheckTicker = time.NewTicker(config.RefreshesCheckInterval()).C

	resSupplier.offersID = 0
	resSupplier.activeOffers = make(map[common.OfferID]*supplierOffer)
	resSupplier.offersMutex = &sync.Mutex{}
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
				sup.offersMutex.Lock()
				defer sup.offersMutex.Unlock()

				var overlayNodes []*overlay.Node

				if !sup.availableResources.IsZero() {
					/*
						- Remove all active offers from the traders in order to concentrate all available resources.
						- This is used to try offer the maximum amount of resources the node has available between
						the: Available (offered) and the Available (but not offered).
					*/
					for offerID, offer := range sup.activeOffers {
						sup.client.RemoveOffer(sup.config.HostIP(), "", offer.ResponsibleTraderIP(),
							offer.ResponsibleTraderGUID().String(), int64(offerID))
						delete(sup.activeOffers, offerID)

						sup.availableResources.Add(*offer.Resources())
					}

					fittestResources := sup.resourcesMap.GetFittestResources(*sup.availableResources)
					destinationGUID, _ := sup.resourcesMap.RandomGUID(*sup.availableResources)
					overlayNodes, _ = sup.overlay.Lookup(destinationGUID.Bytes())
					remoteNodes := sup.getLowerCapacityNodes(overlayNodes, *fittestResources)

					// .. try search nodes in the beginning of the target resource range region
					if len(remoteNodes) == 0 {
						targetResourcesFirstGuid := sup.resourcesMap.FirstGUIDofResources(*fittestResources)
						overlayNodes, _ = sup.overlay.Lookup(targetResourcesFirstGuid.Bytes())
						remoteNodes = sup.getLowerCapacityNodes(overlayNodes, *fittestResources)
					}

					// ... try search for random nodes that handle less powerful resource combinations
					for len(remoteNodes) == 0 {
						destinationGUID, _ = sup.resourcesMap.PreviousRandomGUID(*destinationGUID, *fittestResources)
						if destinationGUID == nil {
							log.Errorf(util.LogTag("Supplier")+"No nodes available to handle resources offer: %s",
								fittestResources.String())
							return // Wait fot the next tick to try supply resources
						}
						overlayNodes, _ = sup.overlay.Lookup(destinationGUID.Bytes())
						remoteNodes = sup.getLowerCapacityNodes(overlayNodes, *fittestResources)
					}

					// Chose the first node returned by the overlay API
					chosenNode := remoteNodes[0]
					chosenNodeGUID := guid.NewGUIDBytes(chosenNode.GUID())

					err := sup.client.CreateOffer(sup.config.HostIP(), "", chosenNode.IP(),
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
						log.Debugf(util.LogTag("Supplier")+"Offer DOWN, OfferID: %d, ResponsibleTraderIP: %s",
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
Remove the overlay nodes that does not represent resources contained inside (lower then) the target resources.
Used during the create offer mechanism in order to select lowest resource combination in order to have at least
some of resources available in the system.
*/
func (sup *Supplier) getLowerCapacityNodes(remoteNodes []*overlay.Node, targetRes resources.Resources) []*overlay.Node {
	resultNodes := make([]*overlay.Node, 0)
	for _, v := range remoteNodes {
		if sup.resourcesMap.IsContainedInTargetResources(*guid.NewGUIDBytes(v.GUID()), targetRes) {
			resultNodes = append(resultNodes, v)
		}
	}
	return resultNodes
}

/*
TODO
*/
func (sup *Supplier) getHigherCapacityNodes(remoteNodes []*overlay.Node, targetRes resources.Resources) []*overlay.Node {
	resultNodes := make([]*overlay.Node, 0)
	for _, v := range remoteNodes {
		if sup.resourcesMap.IsTargetResourcesContained(*guid.NewGUIDBytes(v.GUID()), targetRes) {
			resultNodes = append(resultNodes, v)
		}
	}
	return resultNodes
}

/*
Find a list activeOffers that best suit the target resources given.
*/
func (sup *Supplier) FindOffers(targetResources resources.Resources) []remote.Offer {
	if !sup.isWorking() {
		panic(fmt.Errorf("can't find offers, supplier not working"))
	}

	var destinationGUID *guid.GUID = nil
	numOfTradersContacted := 0
	findPhase := 0
	for {
		var err error = nil

		if numOfTradersContacted >= 7 { // TODO: DeHardcode this value
			return make([]remote.Offer, 0)
		}

		if findPhase == 0 { // Random trader inside resources zone
			destinationGUID, _ = sup.resourcesMap.RandomGUID(targetResources)
		} else if findPhase == 1 { // First trader of the resources zone
			destinationGUID = sup.resourcesMap.FirstGUIDofResources(targetResources)
		} else { // Random trader in higher resources zone
			destinationGUID, err = sup.resourcesMap.NextRandomGUID(*destinationGUID, targetResources)
			if err != nil {
				return make([]remote.Offer, 0)
			}
		}

		overlayNodes, _ := sup.overlay.Lookup(destinationGUID.Bytes())
		remoteNodes := sup.getHigherCapacityNodes(overlayNodes, targetResources)
		for _, node := range remoteNodes {
			err, offers := sup.client.GetOffers(node.IP(), guid.NewGUIDBytes(node.GUID()).String())
			if err == nil && (len(offers) != 0) {
				return offers
			}
			numOfTradersContacted = numOfTradersContacted + 1
		}

		findPhase = findPhase + 1
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
		log.Debugf(util.LogTag("Supplier")+"OfferID: %d refresh FAILED (Offer does not exist)", offerID)
		return false
	}

	if offer.IsResponsibleTrader(*guid.NewGUIDString(fromTraderGUID)) {
		offer.Refresh()
		log.Debugf(util.LogTag("Supplier")+"OfferID: %d refresh SUCCESS", offerID)
		return true
	} else {
		log.Debugf(util.LogTag("Supplier")+"OfferID: %d refresh FAILED (wrong trader)", offerID)
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

		sup.client.RemoveOffer(sup.config.HostIP(), "", supOffer.ResponsibleTraderIP(),
			supOffer.ResponsibleTraderGUID().String(), int64(supOffer.ID()))
		delete(sup.activeOffers, common.OfferID(offerID))

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
