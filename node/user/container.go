package user

import (
	"github.com/strabox/caravela/api/rest"
	"github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/resources"
)

type container struct {
	*common.Container        // Base container
	suppIP            string // IP of the supplier node
}

func newContainer(imageKey string, args []string, portMaps []rest.PortMapping,
	resources resources.Resources, id string, supplierIP string) *container {

	return &container{
		Container: common.NewContainer(imageKey, args, portMaps, resources, id),
		suppIP:    supplierIP,
	}
}

func (cont *container) supplierIP() string {
	return cont.suppIP
}