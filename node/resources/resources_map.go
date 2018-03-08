package resources

import (
	"fmt"
	"github.com/strabox/caravela/node/guid"
)

type ResourcePerc struct {
	Value int
	Percentage int
}

type ResourcesMap struct {
	resourcesIdMap       	[][]*guid.GuidRange	// Matrix of Guid ranges for each resource combination
	cpuIndexes				map[int]int			// Obtain the CPU indexes
	ramIndexes				map[int]int			// Obtain the RAM indexes
	invertCpuIndexes		map[int]int			// Obtain the CPU value
	invertRamIndexes		map[int]int			// Obtain the RAM value
	cpuPartitions			[]int				// The CPU partitions
	ramPartitions			[]int				// The RAM partitions
	numOfCPUPartitions 		int					// Number of CPUs partitions
	numOfRAMPartitions 		int					// Number of RAM partitions
}

func NewResourcesMap(cpuPartitionsPerc []ResourcePerc, ramPartitionsPerc []ResourcePerc) *ResourcesMap {
	rm := &ResourcesMap{}
	rm.numOfCPUPartitions = cap(cpuPartitionsPerc)
	rm.numOfRAMPartitions = cap(ramPartitionsPerc)
	rm.resourcesIdMap = make([][]*guid.GuidRange, rm.numOfCPUPartitions)
	rm.cpuIndexes = make(map[int]int)
	rm.ramIndexes = make(map[int]int)
	rm.invertCpuIndexes = make(map[int]int)
	rm.invertRamIndexes = make(map[int]int)
	rm.cpuPartitions = make([]int, rm.numOfCPUPartitions)
	rm.ramPartitions = make([]int, rm.numOfRAMPartitions)
	
	cpuPartitionsPercentage := make([]int, rm.numOfCPUPartitions)
	ramPartitionsPercentage := make([]int, rm.numOfRAMPartitions)
	
	for i, v := range cpuPartitionsPerc {
		cpuPartitionsPercentage[i] = v.Percentage
		rm.cpuPartitions[i] = v.Value
		rm.cpuIndexes[v.Value] = i
		rm.invertCpuIndexes[i] = v.Value
	}
	
	for i, v := range ramPartitionsPerc {
		ramPartitionsPercentage[i] = v.Percentage
		rm.ramPartitions[i] = v.Value
		rm.ramIndexes[v.Value] = i
		rm.invertRamIndexes[i] = v.Value
	}
	
	cpuBaseGuid := guid.NewGuidInteger(0) // The Guids starts at 0
	cpuPartitions := guid.NewGuidRange(*cpuBaseGuid, *guid.GetMaximumGuid()).CreatePartitions(cpuPartitionsPercentage)
	for partIndex, partValue := range cpuPartitions {
		rm.resourcesIdMap[partIndex] = make([]*guid.GuidRange, rm.numOfRAMPartitions) //Allocate the array of ranges for a CPU and RAM capacity
		rm.resourcesIdMap[partIndex] = partValue.CreatePartitions(ramPartitionsPercentage)
	}

	return rm
}

func (rm *ResourcesMap) GetIndexableResources(resources Resources) *Resources {
	res := &Resources{}
	
	for _, v := range rm.cpuPartitions {
		if resources.CPU() >= v {
			res.SetCPU(v)
		} else {
			break
		}
	}
	
	for _, v := range rm.ramPartitions {
		if resources.RAM() >= v {
			res.SetRAM(v)
		} else {
			break
		}
	}
	
	return res
}

func (rm *ResourcesMap) RandomGuid(resources Resources) (*guid.Guid, error) {
	cpuIndex := rm.cpuIndexes[resources.CPU()]
	ramIndex := rm.ramIndexes[resources.RAM()]
	return rm.resourcesIdMap[cpuIndex][ramIndex].GenerateRandomBetween()
}

func (rm *ResourcesMap) ResourcesByGuid(rGuid guid.Guid) (*Resources, error) {
	for indexCPU, _ := range rm.resourcesIdMap {
		for indexRAM, _ := range rm.resourcesIdMap {
			if rm.resourcesIdMap[indexCPU][indexRAM].Inside(rGuid) {
				return NewResources(rm.invertCpuIndexes[indexCPU], rm.invertRamIndexes[indexRAM]), nil
			}
		}
	}
	return nil, fmt.Errorf("Invalid GUID: %s", rGuid.ToString())
}

func (rm *ResourcesMap) Print() {
	for indexCPU, _ := range rm.resourcesIdMap {
		fmt.Printf("|%v CPUs:", rm.invertCpuIndexes[indexCPU])
		for indexRAM, _ := range rm.resourcesIdMap {
			fmt.Printf(" |%vMB RAM: ", rm.invertRamIndexes[indexRAM])
			rm.resourcesIdMap[indexCPU][indexRAM].Print()
		}
		fmt.Println("")
	}
}
