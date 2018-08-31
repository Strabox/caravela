package supplier

import (
	"context"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/external"
)

type multipleOfferStrategy struct {
	baseOfferStrategy
	updateOffers bool
}

func newMultipleOfferStrategy(config *configuration.Configuration) (OfferingStrategy, error) {
	return &multipleOfferStrategy{
		updateOffers: config.DiscoveryBackend() == "chord-multiple-offer-updates",
		baseOfferStrategy: baseOfferStrategy{
			configs: config,
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
	return m.findOffersLowToHigher(ctx, targetResources)
}

func (m *multipleOfferStrategy) UpdateOffers(availableResources resources.Resources) {
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
		offer, err := m.createAnOffer(int64(m.localSupplier.newOfferID()), resourcePartitionTarget, availableResources)
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
					m.remoteClient.UpdateOffer(
						context.Background(),
						&types.Node{IP: m.configs.HostIP(), GUID: ""},
						&types.Node{IP: suppOffer.ResponsibleTraderIP(), GUID: suppOffer.ResponsibleTraderGUID().String()},
						&types.Offer{
							ID:     int64(suppOffer.ID()),
							Amount: 1,
							Resources: types.Resources{
								CPUClass: types.CPUClass(availableResources.CPUClass()),
								CPUs:     availableResources.CPUs(),
								RAM:      availableResources.RAM(),
							},
						})
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
