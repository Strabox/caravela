package supplier

import (
	"context"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/external"
)

// OffersManager is an interface that can be implemented in an object in order to perform the actions of managing
// the resources offering in the system ina different way.
// This is the fundamental for our initial work, because it is the way to test/compare several approaches.
type OffersManager interface {
	// Init initializes the offers manager structure with the necessary objects, resourcesMap the GUID<->Resource map,
	// overlay the communication node overlay and the remoteClient that allows to communicate with other CARAVELA's nodes.
	Init(resourcesMap *resources.Mapping, overlay external.Overlay, remoteClient external.Caravela)

	// FindOffers searches in the system (with help of the overlay) for node's that have offers that offer at least the
	// same resources as targetResources.
	FindOffers(ctx context.Context, targetResources resources.Resources) []types.AvailableOffer

	// CreateOffer creates a new offer in the system given the current local's available resources.
	CreateOffer(newOfferID int64, availableResources resources.Resources) (*supplierOffer, error)
}
