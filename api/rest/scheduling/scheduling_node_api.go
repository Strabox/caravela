package scheduling

import "github.com/strabox/caravela/api/rest"

type Scheduling interface {
	LaunchContainers(fromBuyerIP string, offerId int64, containerImageKey string, portMappings []rest.PortMapping,
		containerArgs []string, cpus int, ram int) (string, error)
}
