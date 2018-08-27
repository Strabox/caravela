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

func (c *Container) Name() string {
	return c.name
}

func (c *Container) ImageKey() string {
	return c.imageKey
}

func (c *Container) Args() []string {
	return c.args
}

func (c *Container) PortMappings() []types.PortMapping {
	return c.portMaps
}

func (c *Container) Resources() resources.Resources {
	return c.resources
}

func (c *Container) ID() string {
	return c.id
}

func (c *Container) ShortID() string {
	return c.id[:ContainerShortIDSize]
}
