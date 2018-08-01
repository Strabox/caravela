package user

import (
	"github.com/strabox/caravela/api/types"
)

type localScheduler interface {
	SubmitContainers([]types.ContainerConfig) ([]types.ContainerStatus, error)
}
