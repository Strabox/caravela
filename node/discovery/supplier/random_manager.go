package supplier

import (
	"context"
	"errors"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/external"
)

type RandomOffersManager struct {
	configs          *configuration.Configuration
	resourcesMapping *resources.Mapping
	overlay          external.Overlay
	remoteClient     external.Caravela
}

func newRandomOffersManager(config *configuration.Configuration) (OffersManager, error) {
	return &RandomOffersManager{
		configs: config,
	}, nil
}

func (man *RandomOffersManager) Init(resourcesMap *resources.Mapping, overlay external.Overlay, remoteClient external.Caravela) {
	man.resourcesMapping = resourcesMap
	man.overlay = overlay
	man.remoteClient = remoteClient
}

func (man *RandomOffersManager) FindOffers(ctx context.Context, targetResources resources.Resources) []types.AvailableOffer {

	return nil
}

func (man *RandomOffersManager) CreateOffer(newOfferID int64, availableResources resources.Resources) (*supplierOffer, error) {

	return nil, errors.New("impossible advertise offer")
}
