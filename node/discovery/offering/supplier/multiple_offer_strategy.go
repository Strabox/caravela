package supplier

import (
	"context"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/discovery/common"
	"github.com/strabox/caravela/node/external"
)

type MultipleOfferStrategy struct {
	baseOfferStrategy
	updateOffers bool
}

func newMultipleOfferStrategy(config *configuration.Configuration) (OfferingStrategy, error) {
	return &MultipleOfferStrategy{
		updateOffers: config.DiscoveryBackend() == "chord-multiple-offer-updates",
		baseOfferStrategy: baseOfferStrategy{
			configs: config,
		},
	}, nil
}

func (s *MultipleOfferStrategy) Init(supp *Supplier, resourcesMap *resources.Mapping, overlay external.Overlay,
	remoteClient external.Caravela) {
	s.localSupplier = supp
	s.resourcesMapping = resourcesMap
	s.overlay = overlay
	s.remoteClient = remoteClient
}

func (s *MultipleOfferStrategy) FindOffers(ctx context.Context, targetResources resources.Resources) []types.AvailableOffer {
	return s.findOffersLowToHigher(ctx, targetResources)
}

func (s *MultipleOfferStrategy) UpdateOffers(availableResources resources.Resources) {
	lowerPartitions, _ := s.resourcesMapping.LowerPartitionsOffer(availableResources)
	offersToRemove := make([]supplierOffer, 0)

	activeOffers := s.localSupplier.offers()
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
		s.localSupplier.removeOffer(common.OfferID(offerToRemove.ID()))
	}

	for _, toOffer := range lowerPartitions {
		offer, err := s.createAnOffer(int64(s.localSupplier.newOfferID()), toOffer, availableResources)
		if err == nil {
			s.localSupplier.addOffer(offer)
		}
	}

	if s.updateOffers {
		activeOffers := s.localSupplier.offers()
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
