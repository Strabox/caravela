package external

import (
	"context"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
)

// Caravela is the complete API/Interface for the remote client of a node.
type Caravela interface {
	// =============================== Discovery ===============================

	// Sends a create offer message to a trader from a supplier that wants to offer its resources.
	CreateOffer(ctx context.Context, fromNode, toNode *types.Node, offer *types.Offer) error

	// Sends a refresh message from a trader to a supplier. It is used to mutually know that both are alive.
	RefreshOffer(ctx context.Context, fromTrader, toSupp *types.Node, offer *types.Offer) (bool, error)

	// Sends a remove offer message from a supplier to a trader. It means the supplier does not handle the offer anymore.
	RemoveOffer(ctx context.Context, fromSupp, toTrader *types.Node, offer *types.Offer) error

	// Sends a get message to obtain all the available offers in a trader.
	GetOffers(ctx context.Context, fromNode, toTrader *types.Node, relay bool) ([]types.AvailableOffer, error)

	// Sends a message to a neighbor trader saying that a given trader has offers available
	AdvertiseOffersNeighbor(ctx context.Context, fromTrader, toNeighborTrader, traderOffering *types.Node) error

	// =============================== Scheduling ===============================

	// Sends a launch container message to a supplier in order to deploy the container
	LaunchContainer(ctx context.Context, fromBuyer, toSupplier *types.Node, offer *types.Offer,
		containerConfig []types.ContainerConfig) ([]types.ContainerStatus, error)

	// =============================== Containers ===============================

	// Sends a stop container message to a supplier in order to stop the container
	StopLocalContainer(ctx context.Context, toSupplier *types.Node, containerID string) error

	// ============================== Configuration ==============================

	// Sends a message to obtain the system configurations of an existing node. Used by joining nodes to know what are
	// the system configuration parameters and the respective values.
	ObtainConfiguration(ctx context.Context, systemsNode *types.Node) (*configuration.Configuration, error)
}
