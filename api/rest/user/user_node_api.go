package user

import "github.com/strabox/caravela/api/rest"

type User interface {
	SubmitContainers(containerImageKey string, portMappings []rest.PortMapping, containerArgs []string,
		cpus int, ram int) error
	ListContainers() rest.ContainersList
	StopContainers(containersIDs []string) error
	Stop()
}
