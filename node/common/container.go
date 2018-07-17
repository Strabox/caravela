package common

import (
	"github.com/strabox/caravela/api/rest"
	"github.com/strabox/caravela/node/common/resources"
)

// Base structure for a container running in the system.
type Container struct {
	imageKey  string
	args      []string
	portMaps  []rest.PortMapping
	resources resources.Resources

	id string
}

func NewContainer(imageKey string, args []string, portMaps []rest.PortMapping,
	resources resources.Resources, id string) *Container {

	return &Container{
		imageKey: imageKey,
		args:     args,
		portMaps: portMaps,

		resources: resources,
		id:        id,
	}
}

func (cont *Container) ImageKey() string {
	return cont.imageKey
}

func (cont *Container) Args() []string {
	return cont.args
}

func (cont *Container) PortMappings() []rest.PortMapping {
	return cont.portMaps
}

func (cont *Container) Resources() resources.Resources {
	return cont.resources
}

func (cont *Container) ID() string {
	return cont.id
}

func (cont *Container) ShortID() string {
	return cont.id[:12]
}
