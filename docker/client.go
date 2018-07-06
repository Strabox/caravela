package docker

import "github.com/strabox/caravela/api/rest"

/*
Interface for interacting with the Docker daemon.
Provides useful wrappers (for docker API client) for simple interaction with CARAVELA components.
*/
type Client interface {
	/*

	 */
	GetDockerCPUAndRAM() (int, int)

	/*

	 */
	CheckContainerStatus(containerID string) (ContainerStatus, error)

	/*

	 */
	RunContainer(imageKey string, portMappings []rest.PortMapping, args []string, cpus int64,
		ram int) (string, error)

	/*

	 */
	RemoveContainer(containerID string)
}
