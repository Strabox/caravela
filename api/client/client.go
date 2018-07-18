package client

import (
	"github.com/strabox/caravela/api/types"
)

// CARAVELA Golang SDK/client complete interface.
type Client interface {
	// Deploy a container in the system.
	SubmitContainers(containerImageKey string, portMappings []string, arguments []string,
		cpus int, ram int) *Error

	// Stop a container (or set of containers) of the user.
	StopContainers(containersIDs []string) *Error

	// List all the containers the user submitted.
	ListContainers() ([]types.ContainerStatus, *Error)

	// Exits the system, it makes the local node exit.
	Exit() *Error
}
