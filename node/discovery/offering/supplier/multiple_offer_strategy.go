package supplier

import (
	"context"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/external"
	"github.com/strabox/caravela/overlay"
)

type multipleOfferStrategy struct {
	baseOfferStrategy
	updateOffers bool
}

func newMultipleOfferStrategy(node common.Node, config *configuration.Configuration) (OfferingStrategy, error) {
	return &multipleOfferStrategy{
		updateOffers: config.DiscoveryBackend() == "chord-multiple-offer-updates",
		baseOfferStrategy: baseOfferStrategy{
			configs: config,
			node:    node,
		},
	}, nil
}

func (m *multipleOfferStrategy) Init(supp *Supplier, resourcesMapping *resources.Mapping, overlay overlay.Overlay,
	remoteClient external.Caravela) {
	m.localSupplier = supp
	m.resourcesMapping = resourcesMapping
	m.overlay = overlay
	m.remoteClient = remoteClient
}

func (m *multipleOfferStrategy) FindOffers(ctx context.Context, targetResources resources.Resources) []types.AvailableOffer {
	availableOffers := make([]types.AvailableOffer, 0)

	for r := 0; r < m.configs.MaxPartitionsSearch(); r++ {
		destinationGUID, err := m.resourcesMapping.RandGUIDFittestSearch(targetResources)
		if err != nil { // System can't handle that many resources
			return availableOffers
		}

		overlayNodes, _ := m.overlay.Lookup(ctx, destinationGUID.Bytes())
		overlayNodes = m.removeNonTargetNodes(overlayNodes, *destinationGUID)

		for _, node := range overlayNodes {
			offers, err := m.remoteClient.GetOffers(
				ctx,
				&types.Node{}, //TODO: Remove this crap!
				&types.Node{IP: node.IP(), GUID: guid.NewGUIDBytes(node.GUID()).String()},
				true)
			if err == nil && len(offers) != 0 {
				availableOffers = append(availableOffers, offers...)
				break
			}
		}

		if len(availableOffers) > 0 {
			return availableOffers
		}
	}

	return availableOffers
}

func (m *multipleOfferStrategy) UpdateOffers(ctx context.Context, availableResources, usedResources resources.Resources) {
	lowerPartitions, _ := m.resourcesMapping.LowerPartitionsOffer(availableResources)
	offersToRemove := make([]supplierOffer, 0)

	activeOffers := m.localSupplier.offers()
OfferLoop:
	for _, offer := range activeOffers {
		offerPartitionRes := m.resourcesMapping.ResourcesByGUID(*offer.ResponsibleTraderGUID())
		for lp, lowerPartition := range lowerPartitions {
			if offerPartitionRes.Equals(lowerPartition) {
				lowerPartitions = append(lowerPartitions[:lp], lowerPartitions[lp+1:]...)
				continue OfferLoop
			}
		}
		offersToRemove = append(offersToRemove, offer)
	}

	for _, resourcePartitionTarget := range lowerPartitions {
		offer, err := m.createAnOffer(ctx, int64(m.localSupplier.newOfferID()), resourcePartitionTarget, availableResources, usedResources)
		if err == nil {
			m.localSupplier.addOffer(offer)
		}
	}

	for _, offerToRemove := range offersToRemove {
		tmpOfferToRemove := offerToRemove
		removeOffer := func(suppOffer supplierOffer) {
			m.remoteClient.RemoveOffer(
				ctx,
				&types.Node{IP: m.configs.HostIP()},
				&types.Node{IP: suppOffer.ResponsibleTraderIP(), GUID: suppOffer.ResponsibleTraderGUID().String()},
				&types.Offer{ID: int64(suppOffer.ID())})
		}
		if m.configs.Simulation() {
			removeOffer(tmpOfferToRemove)
		} else {
			go removeOffer(tmpOfferToRemove)
		}
		m.localSupplier.removeOffer(tmpOfferToRemove.ID())
	}

	if m.updateOffers {
		activeOffers := m.localSupplier.offers()
		for _, offer := range activeOffers {
			if !offer.Resources().Equals(availableResources) {
				updateOffer := func(suppOffer supplierOffer) {
					err := m.remoteClient.UpdateOffer(
						ctx,
						&types.Node{IP: m.configs.HostIP()},
						&types.Node{IP: suppOffer.ResponsibleTraderIP(), GUID: suppOffer.ResponsibleTraderGUID().String()},
						&types.Offer{
							ID:                int64(suppOffer.ID()),
							Amount:            1,
							ContainersRunning: m.localSupplier.numContainersRunning(),
							FreeResources: types.Resources{
								CPUClass: types.CPUClass(availableResources.CPUClass()),
								CPUs:     availableResources.CPUs(),
								Memory:   availableResources.Memory(),
							},
							UsedResources: types.Resources{
								CPUClass: types.CPUClass(usedResources.CPUClass()),
								CPUs:     usedResources.CPUs(),
								Memory:   usedResources.Memory(),
							},
						})
					m.localSupplier.forceOfferRefresh(offer.ID(), err == nil)
				}

				if m.configs.Simulation() {
					updateOffer(offer)
				} else {
					go updateOffer(offer)
				}
			}
		}
	}
}
