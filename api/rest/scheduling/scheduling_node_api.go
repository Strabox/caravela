package scheduling

import (
	"context"
	"github.com/strabox/caravela/api/types"
)

type Scheduling interface {
	LaunchContainers(ctx context.Context, fromBuyer *types.Node, offer *types.Offer,
		containerConfig []types.ContainerConfig) ([]types.ContainerStatus, error)
}
