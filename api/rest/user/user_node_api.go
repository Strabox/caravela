package user

import (
	"github.com/strabox/caravela/api/types"
)

type User interface {
	SubmitContainers(containerImageKey string, portMappings []types.PortMapping, containerArgs []string,
		cpus int, ram int) error
	ListContainers() []types.ContainerStatus
	StopContainers(containersIDs []string) error
	Stop()
}
