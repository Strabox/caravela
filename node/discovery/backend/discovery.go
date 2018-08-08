package backend

import (
	"context"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
)

type Discovery interface {
	common.Component
	// ========================== Internal Services =============================
	AddTrader(traderGUID guid.GUID)
	FindOffers(ctx context.Context, resources resources.Resources) []types.AvailableOffer
	ObtainResources(offerID int64, resourcesNecessary resources.Resources) bool
	ReturnResources(resources resources.Resources)
	// ======================= External/Remote Services =========================
	CreateOffer(fromNode *types.Node, toNode *types.Node, offer *types.Offer)
	RefreshOffer(fromTrader *types.Node, offer *types.Offer) bool
	RemoveOffer(fromSupp *types.Node, toTrader *types.Node, offer *types.Offer)
	GetOffers(ctx context.Context, fromNode, toTrader *types.Node, relay bool) []types.AvailableOffer
	AdvertiseNeighborOffers(fromTrader, toNeighborTrader, traderOffering *types.Node)
	// ============== External/Remote Services (Only Simulation) ================
	AvailableResourcesSim() types.Resources
	MaximumResourcesSim() types.Resources
	RefreshOffersSim()
	SpreadOffersSim()
}
