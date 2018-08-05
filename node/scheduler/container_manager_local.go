package scheduler

import (
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/node/common/resources"
)

type containerManagerLocal interface {
	StartContainer(fromBuyer *types.Node, offer *types.Offer, containersConfigs []types.ContainerConfig,
		totalResourcesNecessary resources.Resources) ([]types.ContainerStatus, error)
}
