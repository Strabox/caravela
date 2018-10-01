package remote

import (
	"context"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/external"
)

type Client struct {
	httpClient external.Caravela
	clientNode common.Node
}

func NewClient(specificClient external.Caravela, clientNode common.Node) *Client {
	return &Client{
		httpClient: specificClient,
		clientNode: clientNode,
	}
}

func (h *Client) getRequestContext(ctx context.Context) context.Context {
	if h.clientNode != nil {
		ctx = context.WithValue(ctx, types.PartitionsStateKey, h.clientNode.GetSystemPartitionsState().PartitionsState())
		ctx = context.WithValue(ctx, types.NodeGUIDKey, h.clientNode.GUID())
		return ctx
	} else {
		return context.Background()
	}
}

func (h *Client) CreateOffer(ctx context.Context, fromNode, toNode *types.Node, offer *types.Offer) error {
	return h.httpClient.CreateOffer(h.getRequestContext(ctx), fromNode, toNode, offer)
}

func (h *Client) RefreshOffer(ctx context.Context, fromTrader, toSupp *types.Node, offer *types.Offer) (bool, error) {
	return h.httpClient.RefreshOffer(h.getRequestContext(ctx), fromTrader, toSupp, offer)
}

func (h *Client) UpdateOffer(ctx context.Context, fromSupplier, toTrader *types.Node, offer *types.Offer) error {
	return h.httpClient.UpdateOffer(h.getRequestContext(ctx), fromSupplier, toTrader, offer)
}

func (h *Client) RemoveOffer(ctx context.Context, fromSupp, toTrader *types.Node, offer *types.Offer) error {
	return h.httpClient.RemoveOffer(h.getRequestContext(ctx), fromSupp, toTrader, offer)
}

func (h *Client) GetOffers(ctx context.Context, fromNode, toTrader *types.Node, relay bool) ([]types.AvailableOffer, error) {
	return h.httpClient.GetOffers(h.getRequestContext(ctx), fromNode, toTrader, relay)
}

func (h *Client) AdvertiseOffersNeighbor(ctx context.Context, fromTrader, toNeighborTrader, traderOffering *types.Node) error {
	return h.httpClient.AdvertiseOffersNeighbor(h.getRequestContext(ctx), fromTrader, toNeighborTrader, traderOffering)
}

func (h *Client) LaunchContainer(ctx context.Context, fromBuyer, toSupplier *types.Node, offer *types.Offer,
	containersConfigs []types.ContainerConfig) ([]types.ContainerStatus, error) {

	return h.httpClient.LaunchContainer(h.getRequestContext(ctx), fromBuyer, toSupplier, offer, containersConfigs)
}

func (h *Client) StopLocalContainer(ctx context.Context, toSupplier *types.Node, containerID string) error {
	return h.httpClient.StopLocalContainer(h.getRequestContext(ctx), toSupplier, containerID)
}

func (h *Client) ObtainConfiguration(ctx context.Context, systemsNode *types.Node) (*configuration.Configuration, error) {
	return h.httpClient.ObtainConfiguration(h.getRequestContext(ctx), systemsNode)
}
