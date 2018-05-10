package containers

import "github.com/strabox/caravela/node/common/resources"

/*
Represents a container that was submitted in the node.
*/
type Container struct {
	dockerID  string // Container ID (same of the Docker engine)
	buyerIP   string // IP of the node that submitted the container in the system
	resources resources.Resources
}

func NewContainer(dockerID string, buyerIP string, resources resources.Resources) *Container {
	cont := &Container{dockerID, buyerIP, resources}
	return cont
}

func (cont *Container) DockerID() string {
	return cont.dockerID
}

func (cont *Container) BuyerIP() string {
	return cont.buyerIP
}
