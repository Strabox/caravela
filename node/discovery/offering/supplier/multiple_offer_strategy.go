package supplier

import (
	"context"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/external"
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

func (m *multipleOfferStrategy) Init(supp *Supplier, resourcesMapping *resources.Mapping, overlay external.Overlay,
	remoteClient external.Caravela) {
	m.localSupplier = supp
	m.resourcesMapping = resourcesMapping
	m.overlay = overlay
	m.remoteClient = remoteClient
}

func (m *multipleOfferStrategy) FindOffers(ctx context.Context, targetResources resources.Resources) []types.AvailableOffer {
	availableOffers := make([]types.AvailableOffer, 0)

	for r := 0; r < 2; r++ {
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

func (m *multipleOfferStrategy) UpdateOffers(availableResources, usedResources resources.Resources) {
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
		offer, err := m.createAnOffer(int64(m.localSupplier.newOfferID()), resourcePartitionTarget, availableResources, usedResources)
		if err == nil {
			m.localSupplier.addOffer(offer)
		}
	}

	for _, offerToRemove := range offersToRemove {
		removeOffer := func(suppOffer supplierOffer) {
			m.remoteClient.RemoveOffer(
				context.Background(),
				&types.Node{IP: m.configs.HostIP()},
				&types.Node{IP: suppOffer.ResponsibleTraderIP(), GUID: suppOffer.ResponsibleTraderGUID().String()},
				&types.Offer{ID: int64(suppOffer.ID())})
		}
		if m.configs.Simulation() {
			removeOffer(offerToRemove)
		} else {
			go removeOffer(offerToRemove)
		}
		m.localSupplier.removeOffer(offerToRemove.ID())
	}

	if m.updateOffers {
		activeOffers := m.localSupplier.offers()
		for _, offer := range activeOffers {
			if !offer.Resources().Equals(availableResources) {
				updateOffer := func(suppOffer supplierOffer) {
					err := m.remoteClient.UpdateOffer(
						context.Background(),
						&types.Node{IP: m.configs.HostIP()},
						&types.Node{IP: suppOffer.ResponsibleTraderIP(), GUID: suppOffer.ResponsibleTraderGUID().String()},
						&types.Offer{
							ID:     int64(suppOffer.ID()),
							Amount: 1,
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
