package containers

import (
	"github.com/strabox/caravela/node/common/resources"
)

// Represents a container that was submitted to run in a CARAVELA's node.
type Container struct {
	dockerID  string              // Container ID (same of the Docker engine)
	buyerIP   string              // IP of the node that submitted the container in the system TODO: Try use node's GUID and user ID?
	resources resources.Resources // The "real" resources requested by the user to run the container
}

func NewContainer(dockerID string, buyerIP string, resources resources.Resources) *Container {
	return &Container{
		dockerID:  dockerID,
		buyerIP:   buyerIP,
		resources: resources,
	}
}

func (container *Container) DockerID() string {
	return container.dockerID
}

func (container *Container) DockerIDShort() string {
	return container.dockerID
}

func (container *Container) BuyerIP() string {
	return container.buyerIP
}

func (container *Container) Resources() resources.Resources {
	return container.resources
}
