package supplier

import (
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/external"
	"github.com/strabox/caravela/overlay"
)

type OffersManager interface {
	Init(resourcesMap *resources.Mapping, overlay overlay.Overlay, remoteClient external.Caravela)
	FindOffers(targetResources resources.Resources) []types.AvailableOffer
	AdvertiseOffer(newOfferID int64, availableResources resources.Resources) (*supplierOffer, error)
}
