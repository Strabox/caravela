package scheduler

import "github.com/strabox/caravela/api/types"

// Interface that provides the necessary methods to talk with other nodes.
type userRemoteClient interface {
	LaunchContainer(fromBuyer, toSupplier *types.Node, offer *types.Offer, containerConfig []types.ContainerConfig) ([]types.ContainerStatus, error)
	StopLocalContainer(toSupplier *types.Node, containerID string) error
}
