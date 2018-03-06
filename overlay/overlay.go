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
	Create(thisNode local.LocalNode)												// Bootstrap this overlay in the current node
	Join(overlayNodeIP string, overlayNodePort int, thisNode local.LocalNode)	// Join an overlay given a IP and Port of a node that belongs to it
	Lookup(key guid.Guid) []*RemoteNode 									// Get a list of remote nodes
	Leave()																	// Leave the overlay
}
