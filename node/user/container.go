package user

import (
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/resources"
)

type deployedContainer struct {
	*common.Container        // Base container
	suppIP            string // IP of the supplier node
}

func newContainer(name, imageKey string, args []string, portMaps []types.PortMapping,
	resources resources.Resources, id string, supplierIP string) *deployedContainer {

	return &deployedContainer{
		Container: common.NewContainer(name, imageKey, args, portMaps, resources, id),
		suppIP:    supplierIP,
	}
}

func (cont *deployedContainer) supplierIP() string {
	return cont.suppIP
}
