package resources

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewResources(t *testing.T) {
	resources := NewResourcesCPUClass(1, 2, 256)

	assert.Equal(t, 1, resources.CPUClass(), "Invalid CPU Class value!")
	assert.Equal(t, 2, resources.CPUs(), "Invalid CPUs value!")
	assert.Equal(t, 256, resources.Memory(), "Invalid Memory value!")
}

func TestSetCPUClass(t *testing.T) {
	resources := NewResourcesCPUClass(1, 2, 256)

	resources.SetCPUClass(2)

	assert.Equal(t, 2, resources.CPUClass(), "Invalid CPU Class value!")
	assert.Equal(t, 2, resources.CPUs(), "Invalid CPUs value!")
	assert.Equal(t, 256, resources.Memory(), "Invalid Memory value!")
}

func TestSetCPU(t *testing.T) {
	resources := NewResourcesCPUClass(1, 2, 256)

	resources.SetCPUs(4)

	assert.Equal(t, 1, resources.CPUClass(), "Invalid CPU Class value!")
	assert.Equal(t, 4, resources.CPUs(), "Invalid CPUs value!")
	assert.Equal(t, 256, resources.Memory(), "Invalid Memory value!")
}

func TestSetMemory(t *testing.T) {
	resources := NewResourcesCPUClass(1, 2, 256)

	resources.SetMemory(1024)

	assert.Equal(t, 1, resources.CPUClass(), "Invalid CPU Class value!")
	assert.Equal(t, 2, resources.CPUs(), "Invalid CPUs value!")
	assert.Equal(t, 1024, resources.Memory(), "Invalid Memory value!")
}

func TestAddCPU(t *testing.T) {
	resources := NewResourcesCPUClass(2, 2, 256)

	resources.AddCPUs(1)

	assert.Equal(t, 2, resources.CPUClass(), "Invalid CPU Class value!")
	assert.Equal(t, 2+1, resources.CPUs(), "Invalid CPUs value!")
	assert.Equal(t, 256, resources.Memory(), "Invalid Memory value!")
}

func TestAddMemory(t *testing.T) {
	resources := NewResourcesCPUClass(2, 2, 256)

	resources.AddMemory(1024)

	assert.Equal(t, 2, resources.CPUClass(), "Invalid CPU Class value!")
	assert.Equal(t, 2, resources.CPUs(), "Invalid CPUs value!")
	assert.Equal(t, 1024+256, resources.Memory(), "Invalid Memory value!")
}

func TestAdd(t *testing.T) {
	resources := NewResourcesCPUClass(2, 2, 256)
	addResources := NewResourcesCPUClass(1, 2, 256)

	resources.Add(*addResources)

	assert.Equal(t, 2, resources.CPUClass(), "Invalid CPU Class value!")
	assert.Equal(t, 2+2, resources.CPUs(), "Invalid CPUs value!")
	assert.Equal(t, 256+256, resources.Memory(), "Invalid Memory value!")
}

func TestSub_Equals(t *testing.T) {
	resources := NewResourcesCPUClass(2, 2, 256)
	subResources := NewResourcesCPUClass(2, 2, 256)

	resources.Sub(*subResources)

	assert.Equal(t, 2, resources.CPUClass(), "Invalid CPU Class value!")
	assert.Equal(t, 2-2, resources.CPUs(), "Invalid CPUs value!")
	assert.Equal(t, 256-256, resources.Memory(), "Invalid Memory value!")
}

func TestSub_Different(t *testing.T) {
	resources := NewResourcesCPUClass(1, 2, 256)
	subResources := NewResourcesCPUClass(2, 1, 128)

	resources.Sub(*subResources)

	assert.Equal(t, 1, resources.CPUClass(), "Invalid CPU Class value!")
	assert.Equal(t, 2-1, resources.CPUs(), "Invalid CPUs value!")
	assert.Equal(t, 256-128, resources.Memory(), "Invalid Memory value!")
}

func TestSetZero(t *testing.T) {
	resources := NewResourcesCPUClass(2, 2, 256)

	resources.SetZero()

	assert.Equal(t, 2, resources.CPUClass(), "Invalid CPU Class value!")
	assert.Equal(t, 0, resources.CPUs(), "Invalid CPUs value!")
	assert.Equal(t, 0, resources.Memory(), "Invalid Memory value!")
}

func TestSetTo(t *testing.T) {
	resources := NewResourcesCPUClass(1, 2, 256)
	setResources := NewResourcesCPUClass(2, 1, 2046)

	resources.SetTo(*setResources)

	assert.Equal(t, 2, resources.CPUClass(), "Invalid CPU Class value!")
	assert.Equal(t, 1, resources.CPUs(), "Invalid CPUs value!")
	assert.Equal(t, 2046, resources.Memory(), "Invalid Memory value!")
}

func TestIsValid_True(t *testing.T) {
	resources := NewResourcesCPUClass(2, 1, 256)

	res := resources.IsValid()

	assert.True(t, res, "It should have returned true!")
}

func TestIsValid_False_1(t *testing.T) {
	resources := NewResourcesCPUClass(1, 1, 0)

	res := resources.IsValid()

	assert.False(t, res, "It should have returned false!")
}

func TestIsValid_False_2(t *testing.T) {
	resources := NewResourcesCPUClass(2, 0, 1024)

	res := resources.IsValid()

	assert.False(t, res, "It should have returned false!")
}

func TestIsValid_False_3(t *testing.T) {
	resources := NewResourcesCPUClass(1, 0, 0)

	res := resources.IsValid()

	assert.False(t, res, "It should have returned false!")
}

func TestIsZeroTrue(t *testing.T) {
	resources := NewResourcesCPUClass(1, 0, 0)

	res := resources.IsZero()

	assert.True(t, res, "It should have returned true!")
}

func TestIsZeroFalse(t *testing.T) {
	resources := NewResourcesCPUClass(2, 1, 128)

	res := resources.IsZero()

	assert.False(t, res, "It should have returned false!")
}

func TestEqualsTrue_AllEqual(t *testing.T) {
	resources := NewResourcesCPUClass(2, 2, 256)
	resources2 := NewResourcesCPUClass(2, 2, 256)

	res := resources.Equals(*resources2)

	assert.Equal(t, true, res, "It should have returned true!")
}

func TestContainsTrue_HigherCPUClass(t *testing.T) {
	resources := NewResourcesCPUClass(3, 2, 512)
	containedResources := NewResourcesCPUClass(2, 2, 512)

	res := resources.Contains(*containedResources)

	assert.True(t, res, "It should have returned true!")
}

func TestContainsTrue_MoreCPUs(t *testing.T) {
	resources := NewResourcesCPUClass(2, 3, 512)
	containedResources := NewResourcesCPUClass(2, 2, 512)

	res := resources.Contains(*containedResources)

	assert.True(t, res, "It should have returned true!")
}

func TestContainsTrue_MoreMemory(t *testing.T) {
	resources := NewResourcesCPUClass(2, 2, 256)
	containedResources := NewResourcesCPUClass(2, 2, 128)

	res := resources.Contains(*containedResources)

	assert.True(t, res, "It should have returned true!")
}

func TestContainsFalse_DifferentCPUClass(t *testing.T) {
	resources := NewResourcesCPUClass(1, 2, 256)
	containedResources := NewResourcesCPUClass(2, 2, 256)

	res := resources.Contains(*containedResources)

	assert.False(t, res, "It should have returned false!")
}

func TestContainsFalse_CPUGreater(t *testing.T) {
	resources := NewResourcesCPUClass(1, 2, 256)
	containedResources := NewResourcesCPUClass(1, 3, 256)

	res := resources.Contains(*containedResources)

	assert.False(t, res, "It should have returned false!")
}

func TestContainsFalse_MemoryGreater(t *testing.T) {
	resources := NewResourcesCPUClass(2, 2, 256)
	containedResources := NewResourcesCPUClass(2, 2, 512)

	res := resources.Contains(*containedResources)

	assert.Equal(t, false, res, "It should have returned false!")
}

func TestEqualsFalse_LowerCPUCLass(t *testing.T) {
	resources := NewResourcesCPUClass(1, 2, 256)
	resources2 := NewResourcesCPUClass(2, 2, 256)

	res := resources.Equals(*resources2)

	assert.Equal(t, false, res, "It should have returned false!")
}

func TestEqualsFalse_DifferentCPUs(t *testing.T) {
	resources := NewResourcesCPUClass(3, 1, 256)
	resources2 := NewResourcesCPUClass(3, 2, 256)

	res := resources.Equals(*resources2)

	assert.Equal(t, false, res, "It should have returned false!")
}

func TestEqualsFalse_DifferentMemory(t *testing.T) {
	resources := NewResourcesCPUClass(3, 2, 256)
	resources2 := NewResourcesCPUClass(3, 2, 257)

	res := resources.Equals(*resources2)

	assert.Equal(t, false, res, "It should have returned false!")
}

func TestEqualsFalse_AllDifferent(t *testing.T) {
	resources := NewResourcesCPUClass(1, 1, 256)
	resources2 := NewResourcesCPUClass(3, 2, 257)

	res := resources.Equals(*resources2)

	assert.Equal(t, false, res, "It should have returned false!")
}

func TestCopy(t *testing.T) {
	resources := NewResourcesCPUClass(2, 2, 256)

	res := resources.Copy()

	assert.Equal(t, resources.CPUClass(), res.CPUClass(), "CPUClass mismatch!")
	assert.Equal(t, resources.CPUs(), res.CPUs(), "CPUs mismatch!")
	assert.Equal(t, resources.Memory(), res.Memory(), "Memory mismatch!")
}

func TestString(t *testing.T) {
	resources := NewResourcesCPUClass(2, 8, 16384)

	resourcesString := resources.String()

	assert.Equal(t, fmt.Sprintf("<<%d;%d>;%d>", resources.CPUClass(), resources.CPUs(), resources.Memory()), resourcesString, "String mismatch!")
}
