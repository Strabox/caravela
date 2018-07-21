package supplier

import (
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/external"
)

type OffersManager interface {
	Init(resourcesMap *resources.Mapping, overlay external.Overlay, remoteClient external.Caravela)
	FindOffers(targetResources resources.Resources) []types.AvailableOffer
	CreateOffer(newOfferID int64, availableResources resources.Resources) (*supplierOffer, error)
}
