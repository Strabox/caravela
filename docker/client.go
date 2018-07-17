package docker

import (
	"github.com/strabox/caravela/api/rest"
	"github.com/strabox/caravela/docker/container"
)

// Interface for interacting with the Docker daemon.
// Provides a useful wrapper, for docker API client, for simple interaction with CARAVELA components.
type Client interface {
	// Obtains the Docker engine max CPU cores and RAM.
	GetDockerCPUAndRAM() (int, int)

	// Checks the status of a container in the  Docker engine.
	CheckContainerStatus(containerID string) (container.ContainerStatus, error)

	// Runs a container in the Docker engine.
	RunContainer(imageKey string, portMappings []rest.PortMapping, args []string, cpus int64,
		ram int) (string, error)

	// Remove a container from the Docker engine.
	RemoveContainer(containerID string) error
}
