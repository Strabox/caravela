package user

import (
	"context"
	"github.com/strabox/caravela/api/types"
)

type User interface {
	SubmitContainers(ctx context.Context, containersConfigs []types.ContainerConfig) ([]types.ContainerStatus, error)
	ListContainers(ctx context.Context) []types.ContainerStatus
	StopContainers(ctx context.Context, containersIDs []string) error
	Stop(ctx context.Context)
}
