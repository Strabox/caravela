package backend

import (
	"context"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
)

// Discovery is the interface for the resource discovery component of the system.
// Structures that implement this are responsible for:
//     - Managing the node's own resources, co-operating with the local containers manager and scheduler.
//	   - Help the other nodes managing all the resources in teh system implementing a distributed protocol.
// It is not mandatory to implement all the methods listed.
type Discovery interface {
	// Component is the interface that all "runnable" Components in caravela must adhere to.
	common.Component

	//
	GUID() string

	// =========================== Internal Services (Mandatory to Implement) =====================
	//
	AddTrader(traderGUID guid.GUID)
	//
	FindOffers(ctx context.Context, resources resources.Resources) []types.AvailableOffer
	//
	ObtainResources(offerID int64, resourcesNecessary resources.Resources, numContainersToRun int) bool
	//
	ReturnResources(resources resources.Resources, numContainerStopped int)

	// ================================== External/Remote Services ================================
	//
	CreateOffer(fromNode, toNode *types.Node, offer *types.Offer)
	//
	RefreshOffer(fromTrader *types.Node, offer *types.Offer) bool
	//
	UpdateOffer(fromSupp, toTrader *types.Node, offer *types.Offer)
	//
	RemoveOffer(fromSupp, toTrader *types.Node, offer *types.Offer)
	//
	GetOffers(ctx context.Context, fromNode, toTrader *types.Node, relay bool) []types.AvailableOffer
	//
	AdvertiseNeighborOffers(fromTrader, toNeighborTrader, traderOffering *types.Node)

	// ========================== External/Remote Services (Only Simulation) =======================
	//
	NodeInformationSim() (types.Resources, types.Resources, int, int)
	//
	RefreshOffersSim()
	//
	SpreadOffersSim()

	// ===================================== Debug Methods =========================================
	//
	DebugSizeBytes() int
}
