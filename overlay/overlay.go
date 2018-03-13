package overlay

import (
	"github.com/strabox/caravela/node/guid"
	"github.com/strabox/caravela/node/local"
)

/*
Overlay represents an API for a distributed overlay of nodes that allows
us to join it, leave it and lookup for nodes by given Guid
*/
type Overlay interface {
	// Bootstrap this overlay in the current node
	Create(thisNode local.LocalNode)

	// Join an overlay given a IP and Port of a node that belongs to it
	Join(overlayNodeIP string, overlayNodePort int, thisNode local.LocalNode)

	// Get a list of remote nodes
	Lookup(key guid.Guid) []*RemoteNode

	// Leave the overlay
	Leave()
}
