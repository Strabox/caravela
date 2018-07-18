package discovery

import (
	"github.com/strabox/caravela/api/types"
)

// Discovery API necessary to forward the REST calls
type Discovery interface {
	CreateOffer(fromNode *types.Node, toNode *types.Node, offer *types.Offer)
	RefreshOffer(fromTrader *types.Node, offer *types.Offer) bool
	RemoveOffer(fromSupp *types.Node, toTrader *types.Node, offer *types.Offer)
	GetOffers(fromNode, toTrader *types.Node, relay bool) []types.AvailableOffer
	AdvertiseOffersNeighbor(fromTrader, toNeighborTrader, traderOffering *types.Node)
}
