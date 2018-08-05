package scheduler

import (
	"context"
	"github.com/strabox/caravela/api/types"
)

// Interface that provides the necessary methods to talk with other nodes.
type userRemoteClient interface {
	LaunchContainer(ctx context.Context, fromBuyer, toSupplier *types.Node, offer *types.Offer, containerConfig []types.ContainerConfig) ([]types.ContainerStatus, error)
	StopLocalContainer(ctx context.Context, toSupplier *types.Node, containerID string) error
}
