package supplier

import (
	"context"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/external"
	"github.com/strabox/caravela/util"
)

type singleOfferChordStrategy struct {
	baseOfferStrategy
}

func newSingleOfferChordManager(config *configuration.Configuration) (OfferingStrategy, error) {
	return &singleOfferChordStrategy{
		baseOfferStrategy: baseOfferStrategy{
			configs: config,
		},
	}, nil
}

func (s *singleOfferChordStrategy) Init(supp *Supplier, resourcesMap *resources.Mapping, overlay external.Overlay,
	remoteClient external.Caravela) {
	s.localSupplier = supp
	s.resourcesMapping = resourcesMap
	s.overlay = overlay
	s.remoteClient = remoteClient
}

func (s *singleOfferChordStrategy) FindOffers(ctx context.Context, targetResources resources.Resources) []types.AvailableOffer {
	if s.configs.SchedulingPolicy() == "binpack" {
		return s.findOffersLowToHigher(ctx, targetResources)
	} else if s.configs.SchedulingPolicy() == "spread" {
		return s.findOffersHigherToLow(ctx, targetResources)
	} else {
		panic(fmt.Errorf("invalid scheduling policy, %s, for this discovery backend, offering", s.configs.SchedulingPolicy()))
	}
}

func (s *singleOfferChordStrategy) UpdateOffers(availableResources, usedResources resources.Resources) {
	activeOffers := s.localSupplier.offers()

	if len(activeOffers) == 1 {
		activeOffer := activeOffers[0]

		if activeOffer.Resources().Equals(availableResources) {
			return // The active offer has the same resources has the node have now. No need to create other.
		}

		samePartition, _ := s.resourcesMapping.SamePartitionResourcesSearch(availableResources, *s.resourcesMapping.ResourcesByGUID(*activeOffer.ResponsibleTraderGUID()))
		if samePartition { // if the new available resources fit in the same resource partition update the offer and exit.
			updateOffer := func(offer supplierOffer) {
				s.remoteClient.UpdateOffer(
					context.Background(),
					&types.Node{IP: s.configs.HostIP()},
					&types.Node{IP: offer.ResponsibleTraderIP(), GUID: offer.ResponsibleTraderGUID().String()},
					&types.Offer{
						ID:     int64(offer.ID()),
						Amount: 1,
						FreeResources: types.Resources{
							CPUClass: types.CPUClass(availableResources.CPUClass()),
							CPUs:     availableResources.CPUs(),
							RAM:      availableResources.RAM(),
						},
						UsedResources: types.Resources{
							CPUClass: types.CPUClass(usedResources.CPUClass()),
							CPUs:     usedResources.CPUs(),
							RAM:      usedResources.RAM(),
						},
					})
			}
			if s.configs.Simulation() {
				updateOffer(activeOffer) // Send update offer message sequential
			} else {
				go updateOffer(activeOffer) // Send update offer message in background
			}
			return
		}

		removeOffer := func(offer supplierOffer) {
			s.remoteClient.RemoveOffer(
				context.Background(),
				&types.Node{IP: s.configs.HostIP()},
				&types.Node{IP: offer.ResponsibleTraderIP(), GUID: offer.ResponsibleTraderGUID().String()},
				&types.Offer{ID: int64(offer.ID())},
			)
		}
		if s.configs.Simulation() {
			removeOffer(activeOffer) // Send remove offer message sequential
		} else {
			go removeOffer(activeOffer) // Send remove offer message in background
		}
		s.localSupplier.removeOffer(activeOffer.ID())
	} else if len(activeOffers) > 1 {
		panic(errors.New("single offer strategy has more than 1 offer active"))
	}

	newOfferID := s.localSupplier.newOfferID()
	log.Debugf(util.LogTag("SUPPLIER")+"CREATING offer... Offer: %d, Res: <%d;%d>",
		int64(newOfferID), availableResources.CPUs(), availableResources.RAM())

	offer, err := s.createAnOffer(int64(newOfferID), availableResources, availableResources, usedResources)
	if err == nil {
		s.localSupplier.addOffer(offer)
	}
}
