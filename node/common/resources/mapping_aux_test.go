package resources

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestResourcePartitions_CPUPowerPartitions(t *testing.T) {
	resources := ResourcePartitions{}
	resources.cpuPowerPartitions = []CPUPowerPartition{
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

	resPartitions := resources.CPUPowerPartitions()

	expectedPartitions := []ResourcePartition{{Value: 0, Percentage: 50}, {Value: 1, Percentage: 25}, {Value: 2, Percentage: 25}}
	assert.Equal(t, expectedPartitions, resPartitions, "")
}

func TestResourcePartitions_CPUPowerPercentages(t *testing.T) {
	resources := ResourcePartitions{}
	resources.cpuPowerPartitions = []CPUPowerPartition{
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

	resPercentages := resources.CPUPowerPercentages()
	expectedPercentages := []int{50, 25, 25}
	assert.Equal(t, expectedPercentages, resPercentages, "")
}

func TestCPUPowerPartitions_CPUCoresPartitions(t *testing.T) {
	cpuPowerPartition := CPUPowerPartition{}
	cpuPowerPartition.cpuCoresPartitions = []CPUCoresPartition{
		{
			ResourcePartition: ResourcePartition{Value: 1, Percentage: 50},
			ramPartitions:     nil,
		},
		{
			ResourcePartition: ResourcePartition{Value: 2, Percentage: 20},
			ramPartitions:     nil,
		},
		{
			ResourcePartition: ResourcePartition{Value: 4, Percentage: 15},
			ramPartitions:     nil,
		},
		{
			ResourcePartition: ResourcePartition{Value: 8, Percentage: 15},
			ramPartitions:     nil,
		},
	}

	resPartitions := cpuPowerPartition.CPUCoresPartitions()

	expectedPartitions := []ResourcePartition{{Value: 1, Percentage: 50}, {Value: 2, Percentage: 20},
		{Value: 4, Percentage: 15}, {Value: 8, Percentage: 15}}
	assert.Equal(t, expectedPartitions, resPartitions, "")
}

func TestCPUPowerPartitions_CPUCoresPercentages(t *testing.T) {
	cpuPowerPartition := CPUPowerPartition{}
	cpuPowerPartition.cpuCoresPartitions = []CPUCoresPartition{
		{
			ResourcePartition: ResourcePartition{Value: 1, Percentage: 50},
			ramPartitions:     nil,
		},
		{
			ResourcePartition: ResourcePartition{Value: 2, Percentage: 20},
			ramPartitions:     nil,
		},
		{
			ResourcePartition: ResourcePartition{Value: 4, Percentage: 15},
			ramPartitions:     nil,
		},
		{
			ResourcePartition: ResourcePartition{Value: 8, Percentage: 15},
			ramPartitions:     nil,
		},
	}

	resPercentages := cpuPowerPartition.CPUCoresPercentages()

	expectedPercentages := []int{50, 20, 15, 15}
	assert.Equal(t, expectedPercentages, resPercentages, "")
}

func TestCPUCoresPartitions_RAMPartitions(t *testing.T) {
	cpuCoresPartition := CPUCoresPartition{}
	cpuCoresPartition.ramPartitions = []RAMPartition{
		{ResourcePartition: ResourcePartition{Value: 1, Percentage: 25}},
		{ResourcePartition: ResourcePartition{Value: 2, Percentage: 25}},
		{ResourcePartition: ResourcePartition{Value: 4, Percentage: 25}},
		{ResourcePartition: ResourcePartition{Value: 8, Percentage: 25}},
	}

	resPartitions := cpuCoresPartition.RAMPartitions()

	expectedPartitions := []ResourcePartition{{Value: 1, Percentage: 25}, {Value: 2, Percentage: 25},
		{Value: 4, Percentage: 25}, {Value: 8, Percentage: 25}}
	assert.Equal(t, expectedPartitions, resPartitions, "")
}

func TestCPUCoresPartitions_RAMPercentages(t *testing.T) {
	cpuCoresPartition := CPUCoresPartition{}
	cpuCoresPartition.ramPartitions = []RAMPartition{
		{ResourcePartition: ResourcePartition{Value: 1, Percentage: 25}},
		{ResourcePartition: ResourcePartition{Value: 2, Percentage: 25}},
		{ResourcePartition: ResourcePartition{Value: 4, Percentage: 25}},
		{ResourcePartition: ResourcePartition{Value: 8, Percentage: 25}},
	}

	resPercentages := cpuCoresPartition.RAMPercentages()

	expectedPercentages := []int{25, 25, 25, 25}
	assert.Equal(t, expectedPercentages, resPercentages, "")
}
