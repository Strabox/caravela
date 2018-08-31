package common

import (
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/stretchr/testify/assert"
	"testing"
)

const containerNameTest = "Redis"
const imageKeyTest = ""
const containerIDTest = "ASDFAASDJKLASDOIAHDAKJSBDABSMDASDJASDJBASJBDKASDJBASD123ASDNB"
const cpuClass = 1
const cpusTest = 4
const ramTest = 1024

func TestNewContainer(t *testing.T) {
	args := make([]string, 0)
	portMaps := make([]types.PortMapping, 0)
	contResources := *resources.NewResourcesCPUClass(cpuClass, cpusTest, ramTest)
	container := NewContainer(containerNameTest, imageKeyTest, args, portMaps, contResources, containerIDTest)

	assert.Equal(t, containerNameTest, container.Name(), "Container's name is incorrect!")
	assert.Equal(t, imageKeyTest, container.ImageKey(), "Container's image key is incorrect!")
	assert.Equal(t, make([]string, 0), container.Args(), "Container's arguments is incorrect!")
	assert.Equal(t, make([]types.PortMapping, 0), container.PortMappings(), "Container's port mappings is incorrect!")
	assert.Equal(t, contResources, container.Resources(), "Container's resources is incorrect!")
	assert.Equal(t, containerIDTest, container.ID(), "Container's ID is incorrect!")
	assert.Equal(t, containerIDTest[0:ContainerShortIDSize], container.ShortID(), "Container's short ID is incorrect!")
}
