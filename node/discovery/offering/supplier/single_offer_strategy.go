package supplier

import (
	"context"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/discovery/common"
	"github.com/strabox/caravela/node/external"
	"github.com/strabox/caravela/util"
)

type SingleOfferChordStrategy struct {
	baseOfferStrategy
}

func newSingleOfferChordManager(config *configuration.Configuration) (OfferingStrategy, error) {
	return &SingleOfferChordStrategy{
		baseOfferStrategy: baseOfferStrategy{
			configs: config,
		},
	}, nil
}

func (s *SingleOfferChordStrategy) Init(supp *Supplier, resourcesMap *resources.Mapping, overlay external.Overlay,
	remoteClient external.Caravela) {
	s.localSupplier = supp
	s.resourcesMapping = resourcesMap
	s.overlay = overlay
	s.remoteClient = remoteClient
}

func (s *SingleOfferChordStrategy) FindOffers(ctx context.Context, targetResources resources.Resources) []types.AvailableOffer {
	return s.findOffersLowToHigher(ctx, targetResources)
}

func (s *SingleOfferChordStrategy) UpdateOffers(availableResources resources.Resources) {
	// What?: Remove all active offers from the traders in order to gather all available resources.
	// Goal: This is used to try offer the maximum amount of resources the node has available between
	//		 the Available (offered) and the Available (but not offered).
	activeOffers := s.localSupplier.offers()
	for offerID, offer := range activeOffers {
		removeOffer := func(offer supplierOffer) {
			s.remoteClient.RemoveOffer(
				context.Background(),
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
		s.localSupplier.removeOffer(common.OfferID(offerID))
	}

	newOfferID := s.localSupplier.newOfferID()
	log.Debugf(util.LogTag("SUPPLIER")+"CREATING offer... Offer: %d, Res: <%d;%d>",
		int64(newOfferID), availableResources.CPUs(), availableResources.RAM())

	offer, err := s.createAnOffer(int64(newOfferID), availableResources, availableResources)
	if err == nil {
		s.localSupplier.addOffer(offer)
	}
}
