package util

import (
	"github.com/strabox/caravela/api/types"
)

// Create offer struct/JSON used in REST APIs when a supplier offer resources to be used by others.
type CreateOfferMsg struct {
	ToNode   types.Node  `json:"TN"`
	FromNode types.Node  `json:"FN"`
	Offer    types.Offer `json:"O"`
}

// Refresh offer struct/JSON used in remote REST APIs when a trader refresh an offer.
type RefreshOfferMsg struct {
	FromTrader types.Node  `json:"FT"`
	Offer      types.Offer `json:"O"`
}

// UpdateOfferMsg offer struct/JSON used in remote REST APIs when a supplier wants to update its offer.
type UpdateOfferMsg struct {
	FromSupplier types.Node  `json:"FS"`
	ToTrader     types.Node  `json:"TT"`
	Offer        types.Offer `json:"O"`
}

// Response to a refresh offer message used in remote REST APIs when a supplier acknowledges the refresh message
type RefreshOfferResponseMsg struct {
	Refreshed bool `json:"R"` // True if the offer was refreshed succeeded and false otherwise
}

// Remove offer struct/JSON used in remote REST APIs when a supplier remove its offer from a trader.
type OfferRemoveMsg struct {
	FromSupplier types.Node  `json:"FS"`
	ToTrader     types.Node  `json:"TT"`
	Offer        types.Offer `json:"O"`
}

// Get offers struct/JSON used in the REST APIs.
type GetOffersMsg struct {
	FromNode types.Node `json:"FN"`
	ToTrader types.Node `json:"TT"`
	Relay    bool       `json:"R"`
}

// Launch container struct/JSON used in the REST APIs.
type LaunchContainerMsg struct {
	FromBuyer         types.Node              `json:"FB"`
	Offer             types.Offer             `json:"O"`
	ContainersConfigs []types.ContainerConfig `json:"CC"`
}

// Stop container struct/JSON used in the REST APIs
type StopLocalContainerMsg struct {
	ContainerID string `json:"CId"`
}

// Neighbor offer's message struct/JSON used in the REST APIs.
type NeighborOffersMsg struct {
	FromNeighbor     types.Node `json:"FN"`
	ToNeighbor       types.Node `json:"TN"`
	NeighborOffering types.Node `json:"NO"`
}
