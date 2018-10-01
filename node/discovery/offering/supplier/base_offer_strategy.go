package supplier

import (
	"context"
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	nodeCommon "github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/discovery/common"
	"github.com/strabox/caravela/node/external"
	"github.com/strabox/caravela/overlay"
	"github.com/strabox/caravela/util"
)

type localSupplier interface {
	newOfferID() common.OfferID
	addOffer(offer *supplierOffer)
	removeOffer(offerID common.OfferID)
	offers() []supplierOffer
	numContainersRunning() int
	forceOfferRefresh(offerID common.OfferID, success bool)
}

type baseOfferStrategy struct {
	localSupplier    localSupplier
	resourcesMapping *resources.Mapping
	overlay          overlay.Overlay
	remoteClient     external.Caravela
	node             nodeCommon.Node
	configs          *configuration.Configuration
}

func (b *baseOfferStrategy) createAnOffer(ctx context.Context, newOfferID int64, targetResources, realAvailableRes, usedResources resources.Resources) (*supplierOffer, error) {
	var err error
	var overlayNodes []*overlay.OverlayNode = nil

	destinationGUID, err := b.resourcesMapping.RandGUIDOffer(targetResources)
	if err != nil {
		return nil, errors.New("no nodes capable of handle this offer resources")
	}
	overlayNodes, _ = b.overlay.Lookup(ctx, destinationGUID.Bytes())
	overlayNodes = b.removeNonTargetNodes(overlayNodes, *destinationGUID)

	// .. try search nodes in the beginning of the original target resource range region
	if len(overlayNodes) == 0 {
		destinationGUID, err := b.resourcesMapping.FirstGUIDOffer(targetResources)
		if err != nil {
			return nil, err
		}
		overlayNodes, _ = b.overlay.Lookup(ctx, destinationGUID.Bytes())
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
		overlayNodes, _ = b.overlay.Lookup(ctx, destinationGUID.Bytes())
		overlayNodes = b.removeNonTargetNodes(overlayNodes, *destinationGUID)
	}

	// Chose the first node returned by the overlay API
	chosenNode := overlayNodes[0]
	chosenNodeGUID := guid.NewGUIDBytes(chosenNode.GUID())

	err = b.remoteClient.CreateOffer(
		ctx,
		&types.Node{IP: b.configs.HostIP()},
		&types.Node{IP: chosenNode.IP(), GUID: chosenNodeGUID.String()},
		&types.Offer{
			ID:     newOfferID,
			Amount: 1,
			FreeResources: types.Resources{
				CPUClass: types.CPUClass(realAvailableRes.CPUClass()),
				CPUs:     realAvailableRes.CPUs(),
				Memory:   realAvailableRes.Memory(),
			},
			UsedResources: types.Resources{
				CPUClass: types.CPUClass(usedResources.CPUClass()),
				CPUs:     usedResources.CPUs(),
				Memory:   usedResources.Memory(),
			},
		})
	if err == nil {
		return newSupplierOffer(common.OfferID(newOfferID), 1, realAvailableRes, chosenNode.IP(), *chosenNodeGUID), nil
	}

	return nil, errors.New("impossible advertise offer")
}

// Remove nodes that do not belong to that target GUID partition. (Probably because we were target a partition frontier node)
func (b *baseOfferStrategy) removeNonTargetNodes(remoteNodes []*overlay.OverlayNode, targetGuid guid.GUID) []*overlay.OverlayNode {

	resultNodes := make([]*overlay.OverlayNode, 0)
	targetGuidResources := b.resourcesMapping.ResourcesByGUID(targetGuid)
	for _, remoteNode := range remoteNodes {
		remoteNodeResources := b.resourcesMapping.ResourcesByGUID(*guid.NewGUIDBytes(remoteNode.GUID()))
		if targetGuidResources.Equals(*remoteNodeResources) {
			resultNodes = append(resultNodes, remoteNode)
		}
	}
	return resultNodes
}
