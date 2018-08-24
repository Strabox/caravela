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
	"github.com/strabox/caravela/node/external"
	overlayTypes "github.com/strabox/caravela/overlay/types"
	"github.com/strabox/caravela/util"
)

type SingleOfferChordStrategy struct {
	supplier *Supplier

	partitionsState  *SystemResourcePartitions
	configs          *configuration.Configuration
	resourcesMapping *resources.Mapping
	overlay          external.Overlay
	remoteClient     external.Caravela
}

func newSingleOfferChordManager(config *configuration.Configuration) (OfferingStrategy, error) {
	return &SingleOfferChordStrategy{
		partitionsState: NewSystemResourcePartitions(12),
		configs:         config,
	}, nil
}

func (s *SingleOfferChordStrategy) Init(supp *Supplier, resourcesMap *resources.Mapping, overlay external.Overlay,
	remoteClient external.Caravela) {
	s.supplier = supp
	s.resourcesMapping = resourcesMap
	s.overlay = overlay
	s.remoteClient = remoteClient
}

func (s *SingleOfferChordStrategy) UpdatePartitionsState(partitionsState []types.PartitionState) {
	s.partitionsState.MergePartitionsState(partitionsState)
}

func (s *SingleOfferChordStrategy) PartitionsState() []types.PartitionState {
	return s.partitionsState.PartitionsState()
}

func (s *SingleOfferChordStrategy) FindOffers(ctx context.Context, targetResources resources.Resources) []types.AvailableOffer {
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

		if s.partitionsState.Try(targetResPartition) {
			overlayNodes, _ := s.overlay.Lookup(ctx, destinationGUID.Bytes())
			overlayNodes = s.removeNonTargetNodes(overlayNodes, *destinationGUID)

			for _, node := range overlayNodes {
				offers, err := s.remoteClient.GetOffers(
					context.WithValue(ctx, types.PartitionsStateKey, s.partitionsState.PartitionsState()),
					&types.Node{},
					&types.Node{IP: node.IP(), GUID: guid.NewGUIDBytes(node.GUID()).String()},
					true)
				if err == nil && len(offers) != 0 {
					availableOffers = append(availableOffers, offers...)
					s.partitionsState.Hit(targetResPartition)
					break
				} else if err == nil && len(offers) == 0 {
					s.partitionsState.Miss(targetResPartition)
				}
			}

			if len(availableOffers) > 0 {
				return availableOffers
			}
		} else {
			log.Infof(util.LogTag("SUPPLIER")+"SKIPPING Part: <%d,%d>", targetResPartition.CPUs(), targetResPartition.RAM())
		}

		findPhase++
	}
}

func (s *SingleOfferChordStrategy) UpdateOffers(availableResources resources.Resources) {
	// What?: Remove all active offers from the traders in order to gather all available resources.
	// Goal: This is used to try offer the maximum amount of resources the node has available between
	//		 the Available (offered) and the Available (but not offered).
	activeOffers := s.supplier.offers()
	for offerID, offer := range activeOffers {
		removeOffer := func(offer supplierOffer) {
			s.remoteClient.RemoveOffer(
				context.WithValue(context.Background(), types.PartitionsStateKey, s.partitionsState.PartitionsState()),
				&types.Node{IP: s.configs.HostIP(), GUID: ""},
				&types.Node{IP: offer.ResponsibleTraderIP(), GUID: offer.ResponsibleTraderGUID().String()},
				&types.Offer{ID: int64(offer.ID())},
			)
		}
		if s.configs.Simulation() {
			removeOffer(offer) // Send remove offer message sequential
		} else {
			go removeOffer(offer) // Send remove offer message in background
		}
		s.supplier.removeOffer(common.OfferID(offerID))
	}

	newOfferID := s.supplier.newOfferID()
	log.Debugf(util.LogTag("SUPPLIER")+"CREATING offer... Offer: %d, Res: <%d;%d>",
		int64(newOfferID), availableResources.CPUs(), availableResources.RAM())

	offer, err := s.createAnOffer(int64(newOfferID), availableResources)
	if err == nil {
		s.supplier.addOffer(offer)
	}
}

func (s *SingleOfferChordStrategy) createAnOffer(newOfferID int64, availableResources resources.Resources) (*supplierOffer, error) {
	var err error
	var overlayNodes []*overlayTypes.OverlayNode = nil

	destinationGUID, err := s.resourcesMapping.RandGUIDOffer(availableResources)
	if err != nil {
		return nil, errors.New("no nodes capable of handle this offer resources")
	}
	overlayNodes, _ = s.overlay.Lookup(context.Background(), destinationGUID.Bytes())
	overlayNodes = s.removeNonTargetNodes(overlayNodes, *destinationGUID)

	// .. try search nodes in the beginning of the original target resource range region
	if len(overlayNodes) == 0 {
		destinationGUID, err := s.resourcesMapping.FirstGUIDOffer(availableResources)
		if err != nil {
			return nil, err
		}
		overlayNodes, _ = s.overlay.Lookup(context.Background(), destinationGUID.Bytes())
		overlayNodes = s.removeNonTargetNodes(overlayNodes, *destinationGUID)
	}

	// ... try search for random nodes that handle less powerful resource combinations
	for len(overlayNodes) == 0 {
		destinationGUID, err = s.resourcesMapping.LowerRandGUIDOffer(*destinationGUID, availableResources)
		if err != nil {
			log.Errorf(util.LogTag("SUPPLIER")+"NO NODES to handle resources offer: %s. Error: %s",
				availableResources.String(), err)
			return nil, errors.New("no nodes available to accept offer") // Wait fot the next tick to try supply resources
		}
		overlayNodes, _ = s.overlay.Lookup(context.Background(), destinationGUID.Bytes())
		overlayNodes = s.removeNonTargetNodes(overlayNodes, *destinationGUID)
	}

	// Chose the first node returned by the overlay API
	chosenNode := overlayNodes[0]
	chosenNodeGUID := guid.NewGUIDBytes(chosenNode.GUID())

	err = s.remoteClient.CreateOffer(
		context.WithValue(context.Background(), types.PartitionsStateKey, s.partitionsState.PartitionsState()),
		&types.Node{IP: s.configs.HostIP(), GUID: ""},
		&types.Node{IP: chosenNode.IP(), GUID: chosenNodeGUID.String()},
		&types.Offer{
			ID:        newOfferID,
			Amount:    1,
			Resources: types.Resources{CPUs: availableResources.CPUs(), RAM: availableResources.RAM()}})
	if err == nil {
		return newSupplierOffer(common.OfferID(newOfferID), 1, availableResources, chosenNode.IP(), *chosenNodeGUID), nil
	}

	return nil, errors.New("impossible advertise offer")
}

// Remove nodes that do not belong to that target GUID partition. (Probably because we were target a partition frontier node)
func (s *SingleOfferChordStrategy) removeNonTargetNodes(remoteNodes []*overlayTypes.OverlayNode,
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
