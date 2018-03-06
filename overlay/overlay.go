package overlay

import (
	"github.com/strabox/caravela/node"
)

/*
Overlay represents an API for a distributed overlay of nodes that allows
us to join it, leave it and lookup for nodes with given resources in it
*/
type Overlay interface {
	Create()
	Join(overlayNodeIP string, overlayNodePort int)
	Lookup(resources node.Resources) []*node.RemoteNode // Get a list of remote nodes
	Leave()
}
