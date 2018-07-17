package supplier

import (
	"github.com/strabox/caravela/api/remote"
	"github.com/strabox/caravela/node/api"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/overlay"
)

type OffersManager interface {
	Init(resourcesMap *resources.Mapping, overlay overlay.Overlay, remoteClient remote.Caravela)
	FindOffers(targetResources resources.Resources) []api.Offer
	AdvertiseOffer(newOfferID int64, availableResources resources.Resources) (*supplierOffer, error)
}
