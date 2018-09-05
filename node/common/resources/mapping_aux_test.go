package resources

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestResourcePartitions_CPUPowerPartitions(t *testing.T) {
	resources := ResourcePartitions{}
	resources.cpuClassPartitions = []CPUClassPartition{
		{
			ResourcePartition:  ResourcePartition{Value: 0, Percentage: 50},
			cpuCoresPartitions: nil,
		},
		{
			ResourcePartition:  ResourcePartition{Value: 1, Percentage: 25},
			cpuCoresPartitions: nil,
		},
		{
			ResourcePartition:  ResourcePartition{Value: 2, Percentage: 25},
			cpuCoresPartitions: nil,
		},
	}

	resPartitions := resources.CPUClassPartitions()

	expectedPartitions := []ResourcePartition{{Value: 0, Percentage: 50}, {Value: 1, Percentage: 25}, {Value: 2, Percentage: 25}}
	assert.Equal(t, expectedPartitions, resPartitions, "")
}

func TestResourcePartitions_CPUPowerPercentages(t *testing.T) {
	resources := ResourcePartitions{}
	resources.cpuClassPartitions = []CPUClassPartition{
		{
			ResourcePartition:  ResourcePartition{Value: 0, Percentage: 50},
			cpuCoresPartitions: nil,
		},
		{
			ResourcePartition:  ResourcePartition{Value: 1, Percentage: 25},
			cpuCoresPartitions: nil,
		},
		{
			ResourcePartition:  ResourcePartition{Value: 2, Percentage: 25},
			cpuCoresPartitions: nil,
		},
	}

	resPercentages := resources.CPUClassPercentages()
	expectedPercentages := []int{50, 25, 25}
	assert.Equal(t, expectedPercentages, resPercentages, "")
}

func TestCPUPowerPartitions_CPUCoresPartitions(t *testing.T) {
	cpuPowerPartition := CPUClassPartition{}
	cpuPowerPartition.cpuCoresPartitions = []CPUCoresPartition{
		{
			ResourcePartition: ResourcePartition{Value: 1, Percentage: 50},
			memoryPartitions:  nil,
		},
		{
			ResourcePartition: ResourcePartition{Value: 2, Percentage: 20},
			memoryPartitions:  nil,
		},
		{
			ResourcePartition: ResourcePartition{Value: 4, Percentage: 15},
			memoryPartitions:  nil,
		},
		{
			ResourcePartition: ResourcePartition{Value: 8, Percentage: 15},
			memoryPartitions:  nil,
		},
	}

	resPartitions := cpuPowerPartition.CPUCoresPartitions()

	expectedPartitions := []ResourcePartition{{Value: 1, Percentage: 50}, {Value: 2, Percentage: 20},
		{Value: 4, Percentage: 15}, {Value: 8, Percentage: 15}}
	assert.Equal(t, expectedPartitions, resPartitions, "")
}

func TestCPUPowerPartitions_CPUCoresPercentages(t *testing.T) {
	cpuPowerPartition := CPUClassPartition{}
	cpuPowerPartition.cpuCoresPartitions = []CPUCoresPartition{
		{
			ResourcePartition: ResourcePartition{Value: 1, Percentage: 50},
			memoryPartitions:  nil,
		},
		{
			ResourcePartition: ResourcePartition{Value: 2, Percentage: 20},
			memoryPartitions:  nil,
		},
		{
			ResourcePartition: ResourcePartition{Value: 4, Percentage: 15},
			memoryPartitions:  nil,
		},
		{
			ResourcePartition: ResourcePartition{Value: 8, Percentage: 15},
			memoryPartitions:  nil,
		},
	}

	resPercentages := cpuPowerPartition.CPUCoresPercentages()

	expectedPercentages := []int{50, 20, 15, 15}
	assert.Equal(t, expectedPercentages, resPercentages, "")
}

func TestCPUCoresPartitions_MemoryPartitions(t *testing.T) {
	cpuCoresPartition := CPUCoresPartition{}
	cpuCoresPartition.memoryPartitions = []MemoryPartition{
		{ResourcePartition: ResourcePartition{Value: 1, Percentage: 25}},
		{ResourcePartition: ResourcePartition{Value: 2, Percentage: 25}},
		{ResourcePartition: ResourcePartition{Value: 4, Percentage: 25}},
		{ResourcePartition: ResourcePartition{Value: 8, Percentage: 25}},
	}

	resPartitions := cpuCoresPartition.MemoryPartitions()

	expectedPartitions := []ResourcePartition{{Value: 1, Percentage: 25}, {Value: 2, Percentage: 25},
		{Value: 4, Percentage: 25}, {Value: 8, Percentage: 25}}
	assert.Equal(t, expectedPartitions, resPartitions, "")
}

func TestCPUCoresPartitions_MemoryPercentages(t *testing.T) {
	cpuCoresPartition := CPUCoresPartition{}
	cpuCoresPartition.memoryPartitions = []MemoryPartition{
		{ResourcePartition: ResourcePartition{Value: 1, Percentage: 25}},
		{ResourcePartition: ResourcePartition{Value: 2, Percentage: 25}},
		{ResourcePartition: ResourcePartition{Value: 4, Percentage: 25}},
		{ResourcePartition: ResourcePartition{Value: 8, Percentage: 25}},
	}

	resPercentages := cpuCoresPartition.MemoryPercentages()

	expectedPercentages := []int{25, 25, 25, 25}
	assert.Equal(t, expectedPercentages, resPercentages, "")
}
