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

type SmartChordOffersManager struct {
	configs          *configuration.Configuration
	resourcesMapping *resources.Mapping
	overlay          external.Overlay
	remoteClient     external.Caravela
}

func newSmartChordManageOffers(config *configuration.Configuration) (OffersManager, error) {
	return &SmartChordOffersManager{
		configs: config,
	}, nil
}

func (man *SmartChordOffersManager) Init(resourcesMap *resources.Mapping, overlay external.Overlay, remoteClient external.Caravela) {
	man.resourcesMapping = resourcesMap
	man.overlay = overlay
	man.remoteClient = remoteClient
}

func (man *SmartChordOffersManager) FindOffers(ctx context.Context, targetResources resources.Resources) []types.AvailableOffer {
	var destinationGUID *guid.GUID = nil
	findPhase := 0
	availableOffers := make([]types.AvailableOffer, 0)
	for {
		var err error = nil

		if findPhase == 0 { // Random trader inside resources zone
			destinationGUID, err = man.resourcesMapping.RandGUIDSearch(targetResources)
			if err != nil { // System can't handle that many resources
				return availableOffers
			}
		} else { // Random trader in higher resources zone
			destinationGUID, err = man.resourcesMapping.HigherRandGUIDSearch(*destinationGUID, targetResources)
			if err != nil { // No more resource partitions to search
				return availableOffers
			}
		}

		res := man.resourcesMapping.ResourcesByGUID(*destinationGUID)
		log.Debugf(util.LogTag("SUPPLIER")+"FINDING OFFERS %s", res)

		overlayNodes, _ := man.overlay.Lookup(ctx, destinationGUID.Bytes())
		overlayNodes = man.removeNonTargetNodes(overlayNodes, *destinationGUID)

		for _, node := range overlayNodes {
			offers, err := man.remoteClient.GetOffers(
				ctx,
				&types.Node{},
				&types.Node{IP: node.IP(), GUID: guid.NewGUIDBytes(node.GUID()).String()},
				true,
			)
			if err == nil && len(offers) != 0 {
				availableOffers = append(availableOffers, offers...)
				break
			}
		}

		if len(availableOffers) > 0 {
			return availableOffers
		}

		findPhase++
	}
}

func (man *SmartChordOffersManager) CreateOffer(newOfferID int64, availableResources resources.Resources) (*supplierOffer, error) {
	var err error
	var overlayNodes []*overlayTypes.OverlayNode = nil

	destinationGUID, err := man.resourcesMapping.RandGUIDOffer(availableResources)
	if err != nil {
		return nil, errors.New("no nodes to handle offer resources")
	}
	overlayNodes, err = man.overlay.Lookup(context.Background(), destinationGUID.Bytes())
	if err != nil {
		return nil, errors.New("can't publish offer")
	}
	overlayNodes = man.removeNonTargetNodes(overlayNodes, *destinationGUID)

	// .. try search nodes in the beginning of the original target resource range region
	if len(overlayNodes) == 0 {
		destinationGUID, err := man.resourcesMapping.FirstGUIDOffer(availableResources)
		if err != nil {
			return nil, err
		}
		overlayNodes, _ = man.overlay.Lookup(context.Background(), destinationGUID.Bytes())
		overlayNodes = man.removeNonTargetNodes(overlayNodes, *destinationGUID)
	}

	// ... try search for random nodes that handle less powerful resource combinations
	for len(overlayNodes) == 0 {
		destinationGUID, err = man.resourcesMapping.LowerRandGUIDOffer(*destinationGUID, availableResources)
		if err != nil {
			log.Errorf(util.LogTag("SUPPLIER")+"NO NODES to handle resources offer: %s. Error: %s",
				availableResources.String(), err)
			return nil, errors.New("no nodes available to accept offer") // Wait fot the next tick to try supply resources
		}
		overlayNodes, _ = man.overlay.Lookup(context.Background(), destinationGUID.Bytes())
		overlayNodes = man.removeNonTargetNodes(overlayNodes, *destinationGUID)
	}

	// Chose the first node returned by the overlay API
	chosenNode := overlayNodes[0]
	chosenNodeGUID := guid.NewGUIDBytes(chosenNode.GUID())

	err = man.remoteClient.CreateOffer(context.Background(),
		&types.Node{IP: man.configs.HostIP(), GUID: ""},
		&types.Node{IP: chosenNode.IP(), GUID: chosenNodeGUID.String()},
		&types.Offer{
			ID:        newOfferID,
			Amount:    1,
			Resources: types.Resources{CPUs: availableResources.CPUs(), RAM: availableResources.RAM()},
		})
	if err == nil {
		return newSupplierOffer(common.OfferID(newOfferID), 1, availableResources, chosenNode.IP(), *chosenNodeGUID), nil
	}

	return nil, errors.New("impossible advertise offer")
}

// Remove nodes that do not belong to that target GUID partition. (Probably because we were target a partition frontier node)
func (man *SmartChordOffersManager) removeNonTargetNodes(remoteNodes []*overlayTypes.OverlayNode,
	targetGuid guid.GUID) []*overlayTypes.OverlayNode {

	resultNodes := make([]*overlayTypes.OverlayNode, 0)
	targetGuidResources := man.resourcesMapping.ResourcesByGUID(targetGuid)
	for _, remoteNode := range remoteNodes {
		remoteNodeResources := man.resourcesMapping.ResourcesByGUID(*guid.NewGUIDBytes(remoteNode.GUID()))
		if targetGuidResources.Equals(*remoteNodeResources) {
			resultNodes = append(resultNodes, remoteNode)
		}
	}
	return resultNodes
}
