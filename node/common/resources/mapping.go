package resources

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/common/guid"
)

/*
Partition of a given resource through a percentage.
*/
type ResourcePartition struct {
	Value      int
	Percentage int
}

/*
Mapping representation between the GUIDs and resources combinations.
THREAD SAFE, because it is not expected to dynamic mapping of resources.
*/
type Mapping struct {
	resourcesIDMap     [][]*guid.Range // Matrix of GUID ranges for each resource combination
	cpuIndexes         map[int]int     // Obtain the CPUs indexes (given CPUs value)
	ramIndexes         map[int]int     // Obtain the RAM indexes (given RAM value)
	invertCpuIndexes   map[int]int     // Obtain the CPUs value (given CPUs index)
	invertRamIndexes   map[int]int     // Obtain the RAM value (given RAM index)
	cpuPartitions      []int           // The CPUs partitions e.g. (1, 2, 4, ...)
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

	cpuBaseGuid := guid.NewGUIDInteger(0) // The GUID starts at 0
	cpuPartitions := guid.NewGUIDRange(*cpuBaseGuid, *guid.MaximumGUID()).CreatePartitions(cpuPartitionsPercentage)
	for partIndex, partValue := range cpuPartitions {
		// Allocate the array of ranges for a CPUs and RAM capacity
		rm.resourcesIDMap[partIndex] = make([]*guid.Range, rm.numOfRAMPartitions)
		rm.resourcesIDMap[partIndex] = partValue.CreatePartitions(ramPartitionsPercentage)
	}

	return rm
}

func GetCpuCoresPartitions(cpuCoresPartitions []configuration.CPUCoresPartition) []ResourcePartition {
	res := make([]ResourcePartition, len(cpuCoresPartitions))
	for index, partition := range cpuCoresPartitions {
		res[index].Value = partition.Cores
		res[index].Percentage = partition.Percentage
	}
	return res
}

func GetRamPartitions(ramPartitions []configuration.RAMPartition) []ResourcePartition {
	res := make([]ResourcePartition, len(ramPartitions))
	for index, partition := range ramPartitions {
		res[index].Value = partition.RAM
		res[index].Percentage = partition.Percentage
	}
	return res
}

/*
Obtains the fittest resources combination that is contained inside the resources given.
*/
func (rm *Mapping) GetFittestResources(resources Resources) *Resources {
	res := NewResources(0, 0)

	for _, v := range rm.cpuPartitions {
		if resources.CPUs() >= v {
			res.SetCPUs(v)
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
Returns a random GUID in the range of the respective "fittest" target resource combination.
*/
func (rm *Mapping) RandomGUID(targetResources Resources) (*guid.GUID, error) {
	indexesRes := rm.GetFittestResources(targetResources)
	cpuIndex := rm.cpuIndexes[indexesRes.CPUs()]
	ramIndex := rm.ramIndexes[indexesRes.RAM()]
	return rm.resourcesIDMap[cpuIndex][ramIndex].GenerateRandomInside()
}

/*
Returns a random GUID in the next range of resources.
First it tries the GUIDs that represent the SAME cpus and MORE ram.
Second it tries the GUIDs that represent the MORE cpus and SAME ram.
Lastly it will try the GUIDs that represent the MORE cpus and MORE ram.
*/
func (rm *Mapping) NextRandomGUID(currentGUID guid.GUID, targetResources Resources) (*guid.GUID, error) {
	resources, _ := rm.ResourcesByGUID(currentGUID)
	currentCpuIndex := rm.cpuIndexes[resources.CPUs()]
	currentRamIndex := rm.ramIndexes[resources.RAM()]
	if currentRamIndex == (rm.numOfRAMPartitions - 1) {
		if currentCpuIndex == (rm.numOfCPUPartitions - 1) {
			return nil, fmt.Errorf("no more resources combination available")
		} else {
			targetRamIndex := rm.ramIndexes[targetResources.RAM()]
			return rm.resourcesIDMap[currentCpuIndex+1][targetRamIndex].GenerateRandomInside()
		}
	} else {
		return rm.resourcesIDMap[currentCpuIndex][currentRamIndex+1].GenerateRandomInside()
	}
}

/*
Returns a random GUID in the previous range of resources.
First it tries the GUIDs that represent the SAME cpus and LESS ram.
Second it tries the GUIDs that represent the LESS cpus and SAME ram.
Lastly it will try the GUIDs that represent the LESS cpus and LESS ram.
*/
func (rm *Mapping) PreviousRandomGUID(guid guid.GUID, targetResources Resources) (*guid.GUID, error) {
	resources, _ := rm.ResourcesByGUID(guid)
	cpuIndex := rm.cpuIndexes[resources.CPUs()]
	ramIndex := rm.ramIndexes[resources.RAM()]
	if ramIndex == 0 {
		if cpuIndex == 0 {
			return nil, fmt.Errorf("no more resources combination available")
		} else {
			targetRamIndex := rm.ramIndexes[targetResources.RAM()]
			return rm.resourcesIDMap[cpuIndex-1][targetRamIndex].GenerateRandomInside()
		}
	} else {
		return rm.resourcesIDMap[cpuIndex][ramIndex-1].GenerateRandomInside()
	}
}

/*
Verify if the given GUID belongs to a resource combination that is contained inside the
target resource combination.
*/
func (rm *Mapping) IsContainedInTargetResources(guid guid.GUID, targetResources Resources) bool {
	guidResources, _ := rm.ResourcesByGUID(guid)
	return rm.GetFittestResources(targetResources).Contains(*guidResources)
}

/*
TODO
*/
func (rm *Mapping) IsTargetResourcesContained(guid guid.GUID, targetResources Resources) bool {
	guidResources, _ := rm.ResourcesByGUID(guid)
	return guidResources.Contains(*rm.GetFittestResources(targetResources))
}

/*
Returns the first GUID that represents the given resources.
*/
func (rm *Mapping) FirstGUIDofResources(resources Resources) *guid.GUID {
	res := rm.GetFittestResources(resources)
	cpuIndex := rm.cpuIndexes[res.CPUs()]
	ramIndex := rm.ramIndexes[res.RAM()]
	return rm.resourcesIDMap[cpuIndex][ramIndex].LowerGUID()
}

/*
Returns the resources combination that maps to the given GUID.
*/
func (rm *Mapping) ResourcesByGUID(resGUID guid.GUID) (*Resources, error) {
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
	log.Debug("%%%%%%%%%%%%%%%%% Resource <-> GUID Mapping %%%%%%%%%%%%%%%%%%%%%")
	for indexCPU := range rm.resourcesIDMap {
		log.Debugf("-> %v CPUs", rm.invertCpuIndexes[indexCPU])
		for indexRAM := range rm.resourcesIDMap {
			log.Debugf("--> %vMB RAM", rm.invertRamIndexes[indexRAM])
			log.Debug(rm.resourcesIDMap[indexCPU][indexRAM].String())
		}
	}
	log.Debug("%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%")
}
