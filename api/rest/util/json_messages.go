package util

import "github.com/strabox/caravela/api/types"

// Create offer struct/JSON used in REST APIs when a supplier offer resources to be used by others.
type CreateOfferMsg struct {
	ToNode   types.Node  `json:"ToNode"`
	FromNode types.Node  `json:"FromNode"`
	Offer    types.Offer `json:"Offer"`
}

// Refresh offer struct/JSON used in remote REST APIs when a trader refresh an offer.
type RefreshOfferMsg struct {
	FromTrader types.Node  `json:"FromTrader"`
	Offer      types.Offer `json:"Offer"`
}

// Response to a refresh offer message used in remote REST APIs when a supplier acknowledges the refresh message
type RefreshOfferResponseMsg struct {
	Refreshed bool `json:"Refreshed"` // True if the offer was refreshed succeeded and false otherwise
}

// Remove offer struct/JSON used in remote REST APIs when a supplier remove its offer from a trader.
type OfferRemoveMsg struct {
	FromSupplier types.Node  `json:"FromSupplier"`
	ToTrader     types.Node  `json:"ToTrader"`
	Offer        types.Offer `json:"Offer"`
}

// Get offers struct/JSON used in the REST APIs.
type GetOffersMsg struct {
	FromNode types.Node `json:"FromNode"`
	ToTrader types.Node `json:"ToTrader"`
	Relay    bool       `json:"Relay"`
}

// Launch container struct/JSON used in the REST APIs.
type LaunchContainerMsg struct {
	FromBuyer         types.Node              `json:"FromBuyer"`
	Offer             types.Offer             `json:"Offer"`
	ContainersConfigs []types.ContainerConfig `json:"ContainersConfigs"`
}

// Stop container struct/JSON used in the REST APIs
type StopLocalContainerMsg struct {
	ContainerID string `json:"ContainerID"`
}

// Neighbor offer's message struct/JSON used in the REST APIs.
type NeighborOffersMsg struct {
	FromNeighbor     types.Node `json:"FromNeighbor"`
	ToNeighbor       types.Node `json:"ToNeighbor"`
	NeighborOffering types.Node `json:"NeighborOffering"`
}
