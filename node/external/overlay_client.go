package external

import "github.com/strabox/caravela/overlay/types"

// Overlay represents an API for a distributed overlay of nodes that allows
// us to create a new instance of the overlay, join it, leave it, lookup
// for nodes by a given key and get the neighbors of a specific node.
type Overlay interface {
	// Create/Bootstrap the overlay in the current node.
	Create(thisNode types.OverlayMembership) error

	// Join an overlay given a participant node IP and the respective port where its overlay daemon is listening.
	Join(overlayNodeIP string, overlayNodePort int, thisNode types.OverlayMembership) error

	// Get a list of remote nodes using a given key.
	Lookup(key []byte) ([]*types.OverlayNode, error)

	// Get a list of the neighbors nodes of the given node.
	Neighbors(nodeID []byte) ([]*types.OverlayNode, error)

	// Unique identifier for the physical node where the overlay is running.
	// The overlay can have virtual nodes here we only want return always one ID that is unique in all nodes.
	NodeID() ([]byte, error)

	// Leave the overlay.
	Leave() error
}