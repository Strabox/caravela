package supplier

import (
	"context"
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/discovery/common"
	"github.com/strabox/caravela/node/discovery/offering/partitions"
	"github.com/strabox/caravela/node/external"
	overlayTypes "github.com/strabox/caravela/overlay/types"
	"github.com/strabox/caravela/util"
)

type MultipleOfferStrategy struct {
	updateOffers     bool
	supplier         *Supplier
	configs          *configuration.Configuration
	resourcesMapping *resources.Mapping
	overlay          external.Overlay
	remoteClient     external.Caravela
}

func newMultipleOfferStrategy(config *configuration.Configuration) (OfferingStrategy, error) {
	return &MultipleOfferStrategy{
		updateOffers: config.DiscoveryBackend() == "chord-multiple-offer-updates",
		configs:      config,
	}, nil
}

func (s *MultipleOfferStrategy) Init(supp *Supplier, resourcesMap *resources.Mapping, overlay external.Overlay,
	remoteClient external.Caravela) {
	s.supplier = supp
	s.resourcesMapping = resourcesMap
	s.overlay = overlay
	s.remoteClient = remoteClient
}

func (s *MultipleOfferStrategy) FindOffers(ctx context.Context, targetResources resources.Resources) []types.AvailableOffer {
	var destinationGUID *guid.GUID = nil
	findPhase := 0
	availableOffers := make([]types.AvailableOffer, 0)
	for {
		var err error = nil

		if findPhase == 0 { // Random trader inside resources zone
			destinationGUID, err = s.resourcesMapping.RandGUIDSearch(targetResources)
			if err != nil { // System can't handle that many resources
				return availableOffers
			}
		} else { // Random trader in higher resources zone
			destinationGUID, err = s.resourcesMapping.HigherRandGUIDSearch(*destinationGUID, targetResources)
			if err != nil { // No more resource partitions to search
				return availableOffers
			}
		}

		targetResPartition := *s.resourcesMapping.ResourcesByGUID(*destinationGUID)
		log.Debugf(util.LogTag("SUPPLIER")+"FINDING OFFERS for RES: <%d,%d>", targetResPartition.CPUs(), targetResPartition.RAM())

		if partitions.GlobalState.Try(targetResPartition) {
			overlayNodes, _ := s.overlay.Lookup(ctx, destinationGUID.Bytes())
			overlayNodes = s.removeNonTargetNodes(overlayNodes, *destinationGUID)

			for _, node := range overlayNodes {
				offers, err := s.remoteClient.GetOffers(
					ctx,
					&types.Node{},
					&types.Node{IP: node.IP(), GUID: guid.NewGUIDBytes(node.GUID()).String()},
					true)
				if err == nil && len(offers) != 0 {
					availableOffers = append(availableOffers, offers...)
					partitions.GlobalState.Hit(targetResPartition)
					break
				} else if err == nil && len(offers) == 0 {
					partitions.GlobalState.Miss(targetResPartition)
				}
			}

			if len(availableOffers) > 0 {
				return availableOffers
			}
		}

		findPhase++
	}
}

func (s *MultipleOfferStrategy) UpdateOffers(availableResources resources.Resources) {
	lowerPartitions, _ := s.resourcesMapping.LowerPartitionsOffer(availableResources)
	offersToRemove := make([]supplierOffer, 0)

	activeOffers := s.supplier.offers()
OfferLoop:
	for _, offer := range activeOffers {
		offerPartition := s.resourcesMapping.ResourcesByGUID(*offer.ResponsibleTraderGUID())
		for lp, lowerPartition := range lowerPartitions {
			if offerPartition.Equals(lowerPartition) {
				lowerPartitions = append(lowerPartitions[:lp], lowerPartitions[lp+1:]...)
				continue OfferLoop
			}
		}
		offersToRemove = append(offersToRemove, offer)
	}

	for _, offerToRemove := range offersToRemove {
		removeOffer := func(suppOffer supplierOffer) {
			s.remoteClient.RemoveOffer(
				context.Background(),
				&types.Node{IP: s.configs.HostIP(), GUID: ""},
				&types.Node{IP: suppOffer.ResponsibleTraderIP(), GUID: suppOffer.ResponsibleTraderGUID().String()},
				&types.Offer{ID: int64(suppOffer.ID())})
		}
		if s.configs.Simulation() {
			removeOffer(offerToRemove)
		} else {
			go removeOffer(offerToRemove)
		}
		s.supplier.removeOffer(common.OfferID(offerToRemove.ID()))
	}

	for _, toOffer := range lowerPartitions {
		offer, err := s.createAnOffer(int64(s.supplier.newOfferID()), toOffer, availableResources)
		if err == nil {
			s.supplier.addOffer(offer)
		}
	}

	if s.updateOffers {
		activeOffers := s.supplier.offers()
		for _, offer := range activeOffers {
			if !offer.Resources().Equals(availableResources) {
				updateOffer := func(suppOffer supplierOffer) {
					s.remoteClient.UpdateOffer(
						context.Background(),
						&types.Node{IP: s.configs.HostIP(), GUID: ""},
						&types.Node{IP: suppOffer.ResponsibleTraderIP(), GUID: suppOffer.ResponsibleTraderGUID().String()},
						&types.Offer{
							ID:     int64(suppOffer.ID()),
							Amount: 1,
							Resources: types.Resources{
								CPUs: availableResources.CPUs(),
								RAM:  availableResources.RAM(),
							},
						})
				}

				if s.configs.Simulation() {
					updateOffer(offer)
				} else {
					go updateOffer(offer)
				}
			}
		}
	}
}

func (s *MultipleOfferStrategy) createAnOffer(newOfferID int64, targetRes, realAvailableRes resources.Resources) (*supplierOffer, error) {
	var err error
	var overlayNodes []*overlayTypes.OverlayNode = nil

	destinationGUID, err := s.resourcesMapping.RandGUIDOffer(targetRes)
	if err != nil {
		return nil, errors.New("no nodes capable of handle this offer resources")
	}
	overlayNodes, _ = s.overlay.Lookup(context.Background(), destinationGUID.Bytes())
	overlayNodes = s.removeNonTargetNodes(overlayNodes, *destinationGUID)

	// .. try search nodes in the beginning of the original target resource range region
	if len(overlayNodes) == 0 {
		destinationGUID, err := s.resourcesMapping.FirstGUIDOffer(targetRes)
		if err != nil {
			return nil, err
		}
		overlayNodes, _ = s.overlay.Lookup(context.Background(), destinationGUID.Bytes())
		overlayNodes = s.removeNonTargetNodes(overlayNodes, *destinationGUID)
	}

	// ... try search for random nodes that handle less powerful resource combinations
	for len(overlayNodes) == 0 {
		destinationGUID, err = s.resourcesMapping.LowerRandGUIDOffer(*destinationGUID, targetRes)
		if err != nil {
			log.Errorf(util.LogTag("SUPPLIER")+"NO NODES to handle resources offer: %s. Error: %s",
				targetRes.String(), err)
			return nil, errors.New("no nodes available to accept offer") // Wait fot the next tick to try supply resources
		}
		overlayNodes, _ = s.overlay.Lookup(context.Background(), destinationGUID.Bytes())
		overlayNodes = s.removeNonTargetNodes(overlayNodes, *destinationGUID)
	}

	// Chose the first node returned by the overlay API
	chosenNode := overlayNodes[0]
	chosenNodeGUID := guid.NewGUIDBytes(chosenNode.GUID())

	err = s.remoteClient.CreateOffer(
		context.Background(),
		&types.Node{IP: s.configs.HostIP(), GUID: ""},
		&types.Node{IP: chosenNode.IP(), GUID: chosenNodeGUID.String()},
		&types.Offer{
			ID:        newOfferID,
			Amount:    1,
			Resources: types.Resources{CPUs: realAvailableRes.CPUs(), RAM: realAvailableRes.RAM()}})
	if err == nil {
		return newSupplierOffer(common.OfferID(newOfferID), 1, realAvailableRes, chosenNode.IP(), *chosenNodeGUID), nil
	}

	return nil, errors.New("impossible advertise offer")
}

// Remove nodes that do not belong to that target GUID partition. (Probably because we were target a partition frontier node)
func (s *MultipleOfferStrategy) removeNonTargetNodes(remoteNodes []*overlayTypes.OverlayNode,
	targetGuid guid.GUID) []*overlayTypes.OverlayNode {

	resultNodes := make([]*overlayTypes.OverlayNode, 0)
	targetGuidResources := s.resourcesMapping.ResourcesByGUID(targetGuid)
	for _, remoteNode := range remoteNodes {
		remoteNodeResources := s.resourcesMapping.ResourcesByGUID(*guid.NewGUIDBytes(remoteNode.GUID()))
		if targetGuidResources.Equals(*remoteNodeResources) {
			resultNodes = append(resultNodes, remoteNode)
		}
	}
	return resultNodes
}
