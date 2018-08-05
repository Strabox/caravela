package user

import (
	"github.com/strabox/caravela/api/types"
)

type User interface {
	SubmitContainers([]types.ContainerConfig) error
	ListContainers() []types.ContainerStatus
	StopContainers(containersIDs []string) error
	Stop()
}
