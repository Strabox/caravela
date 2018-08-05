package user

import (
	"context"
	"github.com/strabox/caravela/api/types"
)

// Interface that provides the necessary methods to talk with other nodes.
type userRemoteClient interface {
	StopLocalContainer(ctx context.Context, toSupplier *types.Node, containerID string) error
}
