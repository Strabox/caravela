package user

import "github.com/strabox/caravela/api/rest"

type localScheduler interface {
	SubmitContainers(containerImageKey string, portMappings []rest.PortMapping, containerArgs []string,
		cpus int, ram int) (contID string, suppIP string, err error)
}
