package user

import (
	"context"
	"github.com/strabox/caravela/api/types"
)

type localScheduler interface {
	SubmitContainers(ctx context.Context, containersConfigs []types.ContainerConfig) ([]types.ContainerStatus, error)
}
