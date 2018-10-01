package overlay

import (
	"context"
	"github.com/strabox/caravela/api/types"
)

// OverlayClient is the base intermediate structure that interacts with the specific overlay implementations.
type OverlayClient struct {
	specificOverlay Overlay
	localNode       LocalNode
}

func NewOverlayClient(specificOverlay Overlay, localNode LocalNode) *OverlayClient {
	return &OverlayClient{
		specificOverlay: specificOverlay,
		localNode:       localNode,
	}
}

func (o *OverlayClient) getRequestContext(ctx context.Context) context.Context {
	if o.localNode != nil {
		return context.WithValue(ctx, types.NodeGUIDKey, o.localNode.GUID())
	} else {
		return context.Background()
	}
}

/* ============================ Overlay Interface ============================ */

func (o *OverlayClient) Create(ctx context.Context, appNode LocalNode) error {
	return o.specificOverlay.Create(o.getRequestContext(ctx), appNode)
}

func (o *OverlayClient) Join(ctx context.Context, overlayNodeIP string, overlayNodePort int, appNode LocalNode) error {
	return o.specificOverlay.Join(o.getRequestContext(ctx), overlayNodeIP, overlayNodePort, appNode)
}

func (o *OverlayClient) Lookup(ctx context.Context, key []byte) ([]*OverlayNode, error) {
	return o.specificOverlay.Lookup(o.getRequestContext(ctx), key)
}

func (o *OverlayClient) Neighbors(ctx context.Context, nodeID []byte) ([]*OverlayNode, error) {
	return o.specificOverlay.Neighbors(o.getRequestContext(ctx), nodeID)
}

func (o *OverlayClient) NodeID(ctx context.Context) ([]byte, error) {
	return o.specificOverlay.NodeID(o.getRequestContext(ctx))
}

func (o *OverlayClient) Leave(ctx context.Context) error {
	return o.specificOverlay.Leave(o.getRequestContext(ctx))
}
