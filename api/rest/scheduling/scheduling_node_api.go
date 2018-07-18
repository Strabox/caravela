package scheduling

import (
	"github.com/strabox/caravela/api/types"
)

type Scheduling interface {
	LaunchContainers(fromBuyer *types.Node, offer *types.Offer, containerConfig *types.ContainerConfig) (*types.ContainerStatus, error)
}
