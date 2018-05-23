package overlay

import (
	nodeAPI "github.com/strabox/caravela/node/api"
)

/*
Overlay represents an API for a distributed overlay of nodes that allows
us to create a new instance of the overlay, join it, leave it and lookup for nodes by
a given key.
*/
type Overlay interface {
	// Create/Bootstrap the overlay in the current node.
	Create(thisNode nodeAPI.OverlayMembership)

	// Join an overlay given a participant node IP and the respective port where its overlay daemon
	// is listening.
	Join(overlayNodeIP string, overlayNodePort int, thisNode nodeAPI.OverlayMembership)

	// Get a list of remote nodes using a given key.
	Lookup(key []byte) []*Node

	// Leave the overlay.
	Leave()
}
