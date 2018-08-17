package external

import (
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/docker/container"
	"github.com/strabox/caravela/docker/events"
)

// Interface for interacting with the Docker daemon.
// Provides a useful wrapper, for docker API client, for simple interaction with CARAVELA components.
type DockerClient interface {
	Start() <-chan *events.Event

	// Obtains the Docker engine max CPU cores and RAM.
	GetDockerCPUAndRAM() (int, int)

	// Checks the status of a container in the  Docker engine.
	CheckContainerStatus(containerID string) (container.Status, error)

	// Runs a container in the Docker engine.
	RunContainer(contConfig types.ContainerConfig) (*types.ContainerStatus, error)

	// Remove a container from the Docker engine.
	RemoveContainer(containerID string) error
}
