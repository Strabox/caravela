package resources

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewResources(t *testing.T) {
	resources := NewResources(2, 256)

	assert.Equal(t, 2, resources.CPUs(), "Invalid CPUs value!")
	assert.Equal(t, 256, resources.RAM(), "Invalid RAM value!")
}

func TestSetCPU(t *testing.T) {
	resources := NewResources(2, 256)

	resources.SetCPUs(4)

	assert.Equal(t, 4, resources.CPUs(), "Invalid CPUs value!")
	assert.Equal(t, 256, resources.RAM(), "Invalid RAM value!")
}

func TestSetRAM(t *testing.T) {
	resources := NewResources(2, 256)

	resources.SetRAM(1024)

	assert.Equal(t, 2, resources.CPUs(), "Invalid CPUs value!")
	assert.Equal(t, 1024, resources.RAM(), "Invalid RAM value!")
}

func TestAddCPU(t *testing.T) {
	resources := NewResources(2, 256)

	resources.AddCPU(1)

	assert.Equal(t, 2+1, resources.CPUs(), "Invalid CPUs value!")
	assert.Equal(t, 256, resources.RAM(), "Invalid RAM value!")
}

func TestAddRAM(t *testing.T) {
	resources := NewResources(2, 256)

	resources.AddRAM(1024)

	assert.Equal(t, 2, resources.CPUs(), "Invalid CPUs value!")
	assert.Equal(t, 1024+256, resources.RAM(), "Invalid RAM value!")
}

func TestAdd(t *testing.T) {
	resources := NewResources(2, 256)
	addResources := NewResources(2, 256)

	resources.Add(*addResources)

	assert.Equal(t, 2+2, resources.CPUs(), "Invalid CPUs value!")
	assert.Equal(t, 256+256, resources.RAM(), "Invalid RAM value!")
}

func TestSetZero(t *testing.T) {
	resources := NewResources(2, 256)

	resources.SetZero()

	assert.Equal(t, 0, resources.CPUs(), "Invalid CPUs value!")
	assert.Equal(t, 0, resources.RAM(), "Invalid RAM value!")
}

func TestSetTo(t *testing.T) {
	resources := NewResources(2, 256)
	setResources := NewResources(1, 2046)

	resources.SetTo(*setResources)

	assert.Equal(t, 1, resources.CPUs(), "Invalid CPUs value!")
	assert.Equal(t, 2046, resources.RAM(), "Invalid RAM value!")
}

func TestIsZeroTrue(t *testing.T) {
	resources := NewResources(0, 0)

	res := resources.IsZero()

	assert.Equal(t, true, res, "It should have returned true!")
}

func TestIsZeroFalse(t *testing.T) {
	resources := NewResources(1, 128)

	res := resources.IsZero()

	assert.Equal(t, false, res, "It should have returned false!")
}

func TestContainsTrue(t *testing.T) {
	resources := NewResources(2, 256)
	containedResources := NewResources(1, 128)

	res := resources.Contains(*containedResources)

	assert.Equal(t, true, res, "It should have returned true!")
}

func TestContainsFalseCPUGreater(t *testing.T) {
	resources := NewResources(2, 256)
	containedResources := NewResources(3, 256)

	res := resources.Contains(*containedResources)

	assert.Equal(t, false, res, "It should have returned false!")
}

func TestContainsFalseRAMGreater(t *testing.T) {
	resources := NewResources(2, 256)
	containedResources := NewResources(2, 512)

	res := resources.Contains(*containedResources)

	assert.Equal(t, false, res, "It should have returned false!")
}

func TestEqualsTrue(t *testing.T) {
	resources := NewResources(2, 256)
	resources2 := NewResources(2, 256)

	res := resources.Equals(*resources2)

	assert.Equal(t, true, res, "It should have returned true!")
}

func TestEqualsFalse(t *testing.T) {
	resources := NewResources(2, 256)
	resources2 := NewResources(1, 256)

	res := resources.Equals(*resources2)

	assert.Equal(t, false, res, "It should have returned false!")
}

func TestCopy(t *testing.T) {
	resources := NewResources(2, 256)

	res := resources.Copy()

	assert.Equal(t, resources.CPUs(), res.CPUs(), "CPUs mismatch!")
	assert.Equal(t, resources.RAM(), res.RAM(), "RAM mismatch!")
}
