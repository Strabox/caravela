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

type Mapping struct {
	resourcesIdMap     [][]*guid.Range // Matrix of GUID ranges for each resource combination
	cpuIndexes         map[int]int     // Obtain the CPU indexes
	ramIndexes         map[int]int     // Obtain the RAM indexes
	invertCpuIndexes   map[int]int     // Obtain the CPU value
	invertRamIndexes   map[int]int     // Obtain the RAM value
	cpuPartitions      []int           // The CPU partitions
	ramPartitions      []int           // The RAM partitions
	numOfCPUPartitions int             // Number of CPUs partitions
	numOfRAMPartitions int             // Number of RAM partitions
}

/*
Creates a new resource map given the CPUs and RAM partitions and the respective GUID distributions
*/
func NewResourcesMap(cpuPartitionsPerc []ResourcePartition, ramPartitionsPerc []ResourcePartition) *Mapping {
	rm := &Mapping{}
	rm.numOfCPUPartitions = cap(cpuPartitionsPerc)
	rm.numOfRAMPartitions = cap(ramPartitionsPerc)
	rm.resourcesIdMap = make([][]*guid.Range, rm.numOfCPUPartitions)
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
		rm.resourcesIdMap[partIndex] = make([]*guid.Range, rm.numOfRAMPartitions) //Allocate the array of ranges for a CPU and RAM capacity
		rm.resourcesIdMap[partIndex] = partValue.CreatePartitions(ramPartitionsPercentage)
	}

	return rm
}

/*
Obtains the fittest resource combination that is contained inside the resources given
*/
func (rm *Mapping) GetResourcesIndexes(resources Resources) *Resources {
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
Returns a random GUID in the range of the respective resource combination
*/
func (rm *Mapping) RandomGuid(resources Resources) (*guid.Guid, error) {
	indexesRes := rm.GetResourcesIndexes(resources)
	cpuIndex := rm.cpuIndexes[indexesRes.CPU()]
	ramIndex := rm.ramIndexes[indexesRes.RAM()]
	return rm.resourcesIdMap[cpuIndex][ramIndex].GenerateRandomBetween()
}

/*
Returns the resources combination that maps to the given GUID
*/
func (rm *Mapping) ResourcesByGuid(rGuid guid.Guid) (*Resources, error) {
	for indexCPU := range rm.resourcesIdMap {
		for indexRAM := range rm.resourcesIdMap {
			if rm.resourcesIdMap[indexCPU][indexRAM].Inside(rGuid) {
				return NewResources(rm.invertCpuIndexes[indexCPU], rm.invertRamIndexes[indexRAM]), nil
			}
		}
	}
	return nil, fmt.Errorf("invalid GUID %s", rGuid.String())
}

/*
Prints the resource map into the log/std
*/
func (rm *Mapping) Print() {
	log.Infoln("%%%%%%%%%%%%%%%%% Resource <-> GUID Mapping %%%%%%%%%%%%%%%%%%%%%%")
	for indexCPU := range rm.resourcesIdMap {
		log.Infof("-> %v CPUs", rm.invertCpuIndexes[indexCPU])
		for indexRAM := range rm.resourcesIdMap {
			log.Infof("--> %vMB RAM", rm.invertRamIndexes[indexRAM])
			rm.resourcesIdMap[indexCPU][indexRAM].Print()
		}
	}
	log.Infoln("%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%")
}
