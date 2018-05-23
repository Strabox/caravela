package docker

/*
Interface for interacting with the Docker daemon.
Provides useful wrappers for functionality necessary for the CARAVELA.
*/
type Client interface {
	Initialize(runningDockerVersion string)
	GetDockerCPUAndRAM() (int, int)
	CheckContainerStatus(containerID string) (ContainerStatus, error)
	RunContainer(imageKey string, args []string, machineCpus string, ram int) (string, error)
	RemoveContainer(containerID string)
}
