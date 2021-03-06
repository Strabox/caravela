package supplier

import (
	"context"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/external"
	"github.com/strabox/caravela/overlay"
	"github.com/strabox/caravela/util"
)

type singleOfferChordStrategy struct {
	baseOfferStrategy
}

func newSingleOfferChordManager(node common.Node, config *configuration.Configuration) (OfferingStrategy, error) {
	return &singleOfferChordStrategy{
		baseOfferStrategy: baseOfferStrategy{
			configs: config,
			node:    node,
		},
	}, nil
}

func (s *singleOfferChordStrategy) Init(supp *Supplier, resourcesMap *resources.Mapping, overlay overlay.Overlay,
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

func (s *singleOfferChordStrategy) UpdateOffers(ctx context.Context, availableResources, usedResources resources.Resources) {
	activeOffers := s.localSupplier.offers()

	if len(activeOffers) == 1 {
		activeOffer := activeOffers[0]

		if activeOffer.Resources().Equals(availableResources) {
			return // The active offer has the same resources has the node have now. No need to create other.
		}

		samePartition, _ := s.resourcesMapping.SamePartitionResourcesSearch(availableResources, *s.resourcesMapping.ResourcesByGUID(*activeOffer.ResponsibleTraderGUID()))
		if samePartition { // if the new available resources fit in the same resource partition update the offer and exit.
			updateOffer := func(offer supplierOffer) {
				err := s.remoteClient.UpdateOffer(
					ctx,
					&types.Node{IP: s.configs.HostIP()},
					&types.Node{IP: offer.ResponsibleTraderIP(), GUID: offer.ResponsibleTraderGUID().String()},
					&types.Offer{
						ID:                int64(offer.ID()),
						Amount:            1,
						ContainersRunning: s.localSupplier.numContainersRunning(),
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
				s.localSupplier.forceOfferRefresh(offer.ID(), err == nil)
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
				ctx,
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
		int64(newOfferID), availableResources.CPUs(), availableResources.Memory())

	offer, err := s.createAnOffer(ctx, int64(newOfferID), availableResources, availableResources, usedResources)
	if err == nil {
		s.localSupplier.addOffer(offer)
	}
}

func (s *singleOfferChordStrategy) findOffersLowToHigher(ctx context.Context, targetResources resources.Resources) []types.AvailableOffer {
	var destinationGUID *guid.GUID = nil
	findPhase, tries := 0, 0
	availableOffers := make([]types.AvailableOffer, 0)
	for {
		var err error = nil

		if findPhase == 0 { // Random trader inside resources zone
			destinationGUID, err = s.resourcesMapping.RandGUIDFittestSearch(targetResources)
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
		log.Debugf(util.LogTag("SUPPLIER")+"FINDING OFFERS for RES: %s", targetResPartition)

		if s.node.GetSystemPartitionsState().Try(targetResPartition) || !s.configs.SpreadPartitionsState() {
			overlayNodes, _ := s.overlay.Lookup(ctx, destinationGUID.Bytes())
			overlayNodes = s.removeNonTargetNodes(overlayNodes, *destinationGUID)

			for _, node := range overlayNodes {
				offers, err := s.remoteClient.GetOffers(
					ctx,
					&types.Node{}, //TODO: Remove this crap!
					&types.Node{IP: node.IP(), GUID: guid.NewGUIDBytes(node.GUID()).String()},
					true)
				if err == nil && len(offers) != 0 {
					availableOffers = append(availableOffers, offers...)
					s.node.GetSystemPartitionsState().Hit(targetResPartition)
					break
				} else if err == nil && len(offers) == 0 {
					s.node.GetSystemPartitionsState().Miss(targetResPartition)
				}
			}

			if len(availableOffers) > 0 {
				return availableOffers
			}

			tries++
		}

		if tries == (s.configs.MaxPartitionsSearch() - 1) {
			return availableOffers
		}

		findPhase++
	}
}

func (s *singleOfferChordStrategy) findOffersHigherToLow(ctx context.Context, targetResources resources.Resources) []types.AvailableOffer {
	var destinationGUID *guid.GUID = nil
	findPhase, tries := 0, 0
	availableOffers := make([]types.AvailableOffer, 0)
	for {
		var err error = nil

		if findPhase == 0 { // Random trader inside resources zone
			destinationGUID, err = s.resourcesMapping.RandGUIDHighestSearch(targetResources)
			if err != nil { // System can't handle that many resources
				return availableOffers
			}
		} else { // Random trader in higher resources zone
			destinationGUID, err = s.resourcesMapping.LowerRandGUIDSearch(*destinationGUID, targetResources)
			if err != nil { // No more resource partitions to search
				return availableOffers
			}
		}

		targetResPartition := *s.resourcesMapping.ResourcesByGUID(*destinationGUID)
		log.Debugf(util.LogTag("SUPPLIER")+"FINDING OFFERS for RES: %s", targetResPartition)

		if s.node.GetSystemPartitionsState().Try(targetResPartition) || !s.configs.SpreadPartitionsState() {
			overlayNodes, _ := s.overlay.Lookup(ctx, destinationGUID.Bytes())
			overlayNodes = s.removeNonTargetNodes(overlayNodes, *destinationGUID)

			for _, node := range overlayNodes {
				offers, err := s.remoteClient.GetOffers(
					ctx,
					&types.Node{}, //TODO: Remove this crap!
					&types.Node{IP: node.IP(), GUID: guid.NewGUIDBytes(node.GUID()).String()},
					true)
				if err == nil && len(offers) != 0 {
					availableOffers = append(availableOffers, offers...)
					s.node.GetSystemPartitionsState().Hit(targetResPartition)
					break
				} else if err == nil && len(offers) == 0 {
					s.node.GetSystemPartitionsState().Miss(targetResPartition)
				}
			}

			if len(availableOffers) > 0 {
				return availableOffers
			}

			tries++
		}

		if tries == (s.configs.MaxPartitionsSearch() - 1) {
			return availableOffers
		}

		findPhase++
	}
}
