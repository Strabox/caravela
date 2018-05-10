package resources

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/node/common/guid"
)

type ResourcePartition struct {
	Value      int
	Percentage int
}

/*
- Mapping representation between the GUIDs and resources combinations.
- THREAD SAFE, because it is not expected to dynamic mapping of resources.
*/
type Mapping struct {
	resourcesIDMap     [][]*guid.Range // Matrix of GUID ranges for each resource combination
	cpuIndexes         map[int]int     // Obtain the CPU indexes (given CPUs value)
	ramIndexes         map[int]int     // Obtain the RAM indexes (given RAM value)
	invertCpuIndexes   map[int]int     // Obtain the CPU value (given CPUs index)
	invertRamIndexes   map[int]int     // Obtain the RAM value (given RAM index)
	cpuPartitions      []int           // The CPU partitions e.g. (1, 2, 4, ...)
	ramPartitions      []int           // The RAM partitions e.g. (256, 512, 1024, ...)
	numOfCPUPartitions int             // Number of CPUs partitions
	numOfRAMPartitions int             // Number of RAM partitions
}

/*
Creates a new resource map given the CPUs and RAM partitions and the respective GUID distributions.
*/
func NewResourcesMap(cpuPartitionsPerc []ResourcePartition, ramPartitionsPerc []ResourcePartition) *Mapping {
	rm := &Mapping{}
	rm.numOfCPUPartitions = cap(cpuPartitionsPerc)
	rm.numOfRAMPartitions = cap(ramPartitionsPerc)
	rm.resourcesIDMap = make([][]*guid.Range, rm.numOfCPUPartitions)
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

	cpuBaseGuid := guid.NewGuidInteger(0) // The GUID starts at 0
	cpuPartitions := guid.NewGuidRange(*cpuBaseGuid, *guid.MaximumGuid()).CreatePartitions(cpuPartitionsPercentage)
	for partIndex, partValue := range cpuPartitions {
		// Allocate the array of ranges for a CPU and RAM capacity
		rm.resourcesIDMap[partIndex] = make([]*guid.Range, rm.numOfRAMPartitions)
		rm.resourcesIDMap[partIndex] = partValue.CreatePartitions(ramPartitionsPercentage)
	}

	return rm
}

/*
Obtains the fittest resource combination that is contained inside the resources given.
*/
func (rm *Mapping) GetFittestResources(resources Resources) *Resources {
	res := NewResources(0, 0)

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

/*
Returns a random GUID in the range of the respective resource combination.
*/
func (rm *Mapping) RandomGUID(resources Resources) (*guid.Guid, error) {
	indexesRes := rm.GetFittestResources(resources)
	cpuIndex := rm.cpuIndexes[indexesRes.CPU()]
	ramIndex := rm.ramIndexes[indexesRes.RAM()]
	return rm.resourcesIDMap[cpuIndex][ramIndex].GenerateRandomBetween()
}

/*
Returns the resources combination that maps to the given GUID.
*/
func (rm *Mapping) ResourcesByGUID(resGUID guid.Guid) (*Resources, error) {
	for indexCPU := range rm.resourcesIDMap {
		for indexRAM := range rm.resourcesIDMap {
			if rm.resourcesIDMap[indexCPU][indexRAM].Inside(resGUID) {
				return NewResources(rm.invertCpuIndexes[indexCPU], rm.invertRamIndexes[indexRAM]), nil
			}
		}
	}
	return nil, fmt.Errorf("invalid GUID %s", resGUID.String())
}

/*
Prints the resource map into the log.
*/
func (rm *Mapping) Print() {
	log.Debug("%%%%%%%%%%%%%%%%% Resource <-> GUID Mapping %%%%%%%%%%%%%%%%%%%%%%")
	for indexCPU := range rm.resourcesIDMap {
		log.Debugf("-> %v CPUs", rm.invertCpuIndexes[indexCPU])
		for indexRAM := range rm.resourcesIDMap {
			log.Debugf("--> %vMB RAM", rm.invertRamIndexes[indexRAM])
			rm.resourcesIDMap[indexCPU][indexRAM].Print()
		}
	}
	log.Debug("%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%")
}
