package overlay

import (
	nodeAPI "github.com/strabox/caravela/node/api"
)

/*
Overlay represents an API for a distributed overlay of nodes that allows
us to join it, leave it and lookup for nodes by given GUID
*/
type Overlay interface {
	// Bootstrap this overlay in the current node
	Create(thisNode nodeAPI.OverlayMembership)

	// Join an overlay given a IP and Port of a node that belongs to it
	Join(overlayNodeIP string, overlayNodePort int, thisNode nodeAPI.OverlayMembership)

	// Get a list of remote nodes
	Lookup(key []byte) []*Node

	// Leave the overlay
	Leave()
}
