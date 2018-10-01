package overlay

import (
	"context"
)

// Overlay represents an API for a distributed overlay of nodes that allows
// us to create a new instance of the overlay, join it, leave it, lookup
// for nodes by a given key and get the neighbors of a specific node.
type Overlay interface {
	// Create/Bootstrap the overlay in the current node.
	Create(ctx context.Context, thisNode LocalNode) error

	// Join an overlay given a participant node IP and the respective port where its overlay daemon is listening.
	Join(ctx context.Context, overlayNodeIP string, overlayNodePort int, thisNode LocalNode) error

	// Get a list of remote nodes using a given key.
	Lookup(ctx context.Context, key []byte) ([]*OverlayNode, error)

	// Get a list of the neighbors nodes of the given node.
	Neighbors(ctx context.Context, nodeID []byte) ([]*OverlayNode, error)

	// Unique identifier for the physical node where the overlay is running.
	// The overlay can have virtual nodes here we only want return always one ID that is unique in all nodes.
	NodeID(ctx context.Context) ([]byte, error)

	// Leave the overlay.
	Leave(ctx context.Context) error
}
