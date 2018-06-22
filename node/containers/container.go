package containers

import "github.com/strabox/caravela/node/common/resources"

/*
Represents a container that was submitted into a CARAVELA's node.
*/
type Container struct {
	dockerID  string              // Container ID (same of the Docker engine) TODO: Define a different ID at CARAVELA level ??
	buyerIP   string              // IP of the node that submitted the container in the system TODO: Try use node's GUID
	resources resources.Resources // The "real" resources asked by the user to run the container
}

func NewContainer(dockerID string, buyerIP string, resources resources.Resources) *Container {
	cont := &Container{dockerID: dockerID, buyerIP: buyerIP, resources: resources}
	return cont
}

func (container *Container) DockerID() string {
	return container.dockerID
}

func (container *Container) BuyerIP() string {
	return container.buyerIP
}

func (container *Container) Resources() resources.Resources {
	return container.resources
}
