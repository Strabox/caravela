package user

import "github.com/strabox/caravela/api/types"

// Interface that provides the necessary methods to talk with other nodes.
type userRemoteClient interface {
	StopLocalContainer(toSupplier *types.Node, containerID string) error
}
