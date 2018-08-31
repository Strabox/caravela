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

type localSupplier interface {
	newOfferID() common.OfferID
	addOffer(offer *supplierOffer)
	removeOffer(offerID common.OfferID)
	offers() []supplierOffer
}

type baseOfferStrategy struct {
	localSupplier    localSupplier
	resourcesMapping *resources.Mapping
	overlay          external.Overlay
	remoteClient     external.Caravela
	configs          *configuration.Configuration
}

func (b *baseOfferStrategy) findOffersLowToHigher(ctx context.Context, targetResources resources.Resources) []types.AvailableOffer {
	var destinationGUID *guid.GUID = nil
	findPhase := 0
	availableOffers := make([]types.AvailableOffer, 0)
	for {
		var err error = nil

		if findPhase == 0 { // Random trader inside resources zone
			destinationGUID, err = b.resourcesMapping.RandGUIDSearch(targetResources)
			if err != nil { // System can't handle that many resources
				return availableOffers
			}
		} else { // Random trader in higher resources zone
			destinationGUID, err = b.resourcesMapping.HigherRandGUIDSearch(*destinationGUID, targetResources)
			if err != nil { // No more resource partitions to search
				return availableOffers
			}
		}

		targetResPartition := *b.resourcesMapping.ResourcesByGUID(*destinationGUID)
		log.Debugf(util.LogTag("SUPPLIER")+"FINDING OFFERS for RES: %s", targetResPartition)

		if partitions.GlobalState.Try(targetResPartition) {
			overlayNodes, _ := b.overlay.Lookup(ctx, destinationGUID.Bytes())
			overlayNodes = b.removeNonTargetNodes(overlayNodes, *destinationGUID)

			for _, node := range overlayNodes {
				offers, err := b.remoteClient.GetOffers(
					ctx,
					&types.Node{}, //TODO: Remove this crap!
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

func (b *baseOfferStrategy) createAnOffer(newOfferID int64, targetResources, realAvailableRes resources.Resources) (*supplierOffer, error) {
	var err error
	var overlayNodes []*overlayTypes.OverlayNode = nil

	destinationGUID, err := b.resourcesMapping.RandGUIDOffer(targetResources)
	if err != nil {
		return nil, errors.New("no nodes capable of handle this offer resources")
	}
	overlayNodes, _ = b.overlay.Lookup(context.Background(), destinationGUID.Bytes())
	overlayNodes = b.removeNonTargetNodes(overlayNodes, *destinationGUID)

	// .. try search nodes in the beginning of the original target resource range region
	if len(overlayNodes) == 0 {
		destinationGUID, err := b.resourcesMapping.FirstGUIDOffer(targetResources)
		if err != nil {
			return nil, err
		}
		overlayNodes, _ = b.overlay.Lookup(context.Background(), destinationGUID.Bytes())
		overlayNodes = b.removeNonTargetNodes(overlayNodes, *destinationGUID)
	}

	// ... try search for random nodes that handle less powerful resource combinations
	for len(overlayNodes) == 0 {
		destinationGUID, err = b.resourcesMapping.LowerRandGUIDOffer(*destinationGUID, targetResources)
		if err != nil {
			log.Errorf(util.LogTag("SUPPLIER")+"NO NODES to handle resources offer: %s. Error: %s",
				targetResources.String(), err)
			return nil, errors.New("no nodes available to accept offer") // Wait fot the next tick to try supply resources
		}
		overlayNodes, _ = b.overlay.Lookup(context.Background(), destinationGUID.Bytes())
		overlayNodes = b.removeNonTargetNodes(overlayNodes, *destinationGUID)
	}

	// Chose the first node returned by the overlay API
	chosenNode := overlayNodes[0]
	chosenNodeGUID := guid.NewGUIDBytes(chosenNode.GUID())

	err = b.remoteClient.CreateOffer(
		context.Background(),
		&types.Node{IP: b.configs.HostIP(), GUID: ""},
		&types.Node{IP: chosenNode.IP(), GUID: chosenNodeGUID.String()},
		&types.Offer{
			ID:     newOfferID,
			Amount: 1,
			Resources: types.Resources{
				CPUClass: types.CPUClass(realAvailableRes.CPUClass()),
				CPUs:     realAvailableRes.CPUs(),
				RAM:      realAvailableRes.RAM(),
			},
		})
	if err == nil {
		return newSupplierOffer(common.OfferID(newOfferID), 1, realAvailableRes, chosenNode.IP(), *chosenNodeGUID), nil
	}

	return nil, errors.New("impossible advertise offer")
}

// Remove nodes that do not belong to that target GUID partition. (Probably because we were target a partition frontier node)
func (b *baseOfferStrategy) removeNonTargetNodes(remoteNodes []*overlayTypes.OverlayNode, targetGuid guid.GUID) []*overlayTypes.OverlayNode {

	resultNodes := make([]*overlayTypes.OverlayNode, 0)
	targetGuidResources := b.resourcesMapping.ResourcesByGUID(targetGuid)
	for _, remoteNode := range remoteNodes {
		remoteNodeResources := b.resourcesMapping.ResourcesByGUID(*guid.NewGUIDBytes(remoteNode.GUID()))
		if targetGuidResources.Equals(*remoteNodeResources) {
			resultNodes = append(resultNodes, remoteNode)
		}
	}
	return resultNodes
}
