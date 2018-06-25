package docker

/*
Interface for interacting with the Docker daemon.
Provides useful wrappers (for docker API) for functionality necessary to the CARAVELA.
*/
type Client interface {
	GetDockerCPUAndRAM() (int, int)
	CheckContainerStatus(containerID string) (ContainerStatus, error)
	RunContainer(imageKey string, args []string, machineCpus string, ram int) (string, error)
	RemoveContainer(containerID string)
}
