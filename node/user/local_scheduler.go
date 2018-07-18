package user

import (
	"github.com/strabox/caravela/api/types"
)

type localScheduler interface {
	SubmitContainers(containerImageKey string, portMappings []types.PortMapping, containerArgs []string,
		cpus int, ram int) (contID string, suppIP string, err error)
}
