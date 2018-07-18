package containers

import (
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/resources"
)

// Represents a container that was submitted to run in a CARAVELA's node.
type localContainer struct {
	*common.Container // Base container

	buyerIP string // IP of the node that submitted the container in the system TODO: Try use node's GUID and user ID?
}

func newContainer(imageKey string, args []string, portMaps []types.PortMapping, resources resources.Resources,
	dockerID string, buyerIP string) *localContainer {
	return &localContainer{
		Container: common.NewContainer(imageKey, args, portMaps, resources, dockerID),
		buyerIP:   buyerIP,
	}
}

func (container *localContainer) BuyerIP() string {
	return container.buyerIP
}
