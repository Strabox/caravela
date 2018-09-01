package swarm

import (
	"context"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/external"
)

type Discovery struct {
	common.NodeComponent // Base component

	config  *configuration.Configuration // System's configurations.
	overlay external.Overlay             // Overlay component.
	client  external.Caravela            // Remote caravela's client.

	resourcesMap *resources.Mapping // GUID<->Resources mapping
}

// ====================== Local Services (Consumed by other Components) ============================

// Adds a new local "virtual" trader when the overlay notifies its presence.
func (d *Discovery) AddTrader(traderGUID guid.GUID) {
	// Do Nothing - Not necessary for this backend.
}

func (d *Discovery) FindOffers(ctx context.Context, resources resources.Resources) []types.AvailableOffer {
	// Do Nothing - Not necessary for this backend.
	return nil
}

func (d *Discovery) ObtainResources(offerID int64, resourcesNecessary resources.Resources) bool {
	// Do Nothing - Not necessary for this backend.
	return false
}

func (d *Discovery) ReturnResources(resources resources.Resources) {
	// Do Nothing - Not necessary for this backend.
}

// ======================= External Services (Consumed by other Nodes) ==============================

func (d *Discovery) CreateOffer(fromSupp *types.Node, toTrader *types.Node, offer *types.Offer) {
	// Do Nothing - Not necessary for this backend.
}

func (d *Discovery) RefreshOffer(fromTrader *types.Node, offer *types.Offer) bool {
	// Do Nothing - Not necessary for this backend.
	return false
}

func (d *Discovery) UpdateOffer(fromSupp, toTrader *types.Node, offer *types.Offer) {
	// Do Nothing - Not necessary for this backend.
}

func (d *Discovery) RemoveOffer(fromSupp *types.Node, toTrader *types.Node, offer *types.Offer) {
	// Do Nothing - Not necessary for this backend.
}

func (d *Discovery) GetOffers(ctx context.Context, fromNode, toTrader *types.Node, relay bool) []types.AvailableOffer {
	// Do Nothing - Not necessary for this backend.
	return nil
}

func (d *Discovery) AdvertiseNeighborOffers(fromTrader, toNeighborTrader, traderOffering *types.Node) {
	// Do Nothing - Not necessary for this backend.
}

// ======================= External Services (Consumed during simulation ONLY) =========================

// Simulation
func (d *Discovery) AvailableResourcesSim() types.Resources {
	// Do Nothing - Not necessary for this backend.
	return types.Resources{} //TODO
}

// Simulation
func (d *Discovery) MaximumResourcesSim() types.Resources {
	// Do Nothing - Not necessary for this backend.
	return types.Resources{} //TODO
}

// Simulation
func (d *Discovery) RefreshOffersSim() {
	// Do Nothing - Not necessary for this backend.
}

// Simulation
func (d *Discovery) SpreadOffersSim() {
	// Do Nothing - Not necessary for this backend.
}

// ===============================================================================
// =							SubComponent Interface                           =
// ===============================================================================

func (d *Discovery) Start() {
	d.Started(d.config.Simulation(), func() {
		// Do Nothing
	})
}

func (d *Discovery) Stop() {
	d.Stopped(func() {
		// Do Nothing
	})
}

func (d *Discovery) IsWorking() bool {
	return d.Working()
}
