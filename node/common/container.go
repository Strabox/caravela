package common

import (
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/node/common/resources"
)

const ContainerShortIDSize = 12

// Container base structure for a container running in the system.
type Container struct {
	name      string
	imageKey  string
	args      []string
	portMaps  []types.PortMapping
	resources resources.Resources

	id string
}

func NewContainer(name, imageKey string, args []string, portMaps []types.PortMapping,
	resources resources.Resources, id string) *Container {

	return &Container{
		name:     name,
		imageKey: imageKey,
		args:     args,
		portMaps: portMaps,

		resources: resources,
		id:        id,
	}
}

func (cont *Container) Name() string {
	return cont.name
}

func (cont *Container) ImageKey() string {
	return cont.imageKey
}

func (cont *Container) Args() []string {
	return cont.args
}

func (cont *Container) PortMappings() []types.PortMapping {
	return cont.portMaps
}

func (cont *Container) Resources() resources.Resources {
	return cont.resources
}

func (cont *Container) ID() string {
	return cont.id
}

func (cont *Container) ShortID() string {
	return cont.id[:ContainerShortIDSize]
}
