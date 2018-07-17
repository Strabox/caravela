package user

import (
	"github.com/strabox/caravela/api/rest"
	"github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/resources"
)

type deployedContainer struct {
	*common.Container        // Base container
	suppIP            string // IP of the supplier node
}

func newContainer(imageKey string, args []string, portMaps []rest.PortMapping,
	resources resources.Resources, id string, supplierIP string) *deployedContainer {

	return &deployedContainer{
		Container: common.NewContainer(imageKey, args, portMaps, resources, id),
		suppIP:    supplierIP,
	}
}

func (cont *deployedContainer) supplierIP() string {
	return cont.suppIP
}
