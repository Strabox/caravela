package node

import (
	"fmt"
	"math"
)

type ResourcesMap struct {
	resourcesIdMap       [][]*GuidRange
	numOfCPUCombinations int
	numOfRAMCombinations int
	minNumOfCPU          int
	minAmountOfRAM       int
}

const MINIMUM_NUMBER_OF_CPUS = 1  // Must always be POWERS OF TWO
const MINIMUM_AMOUNT_OF_RAM = 128 // Must always be POWERS OF TWO

func NewResourcesMap(cpuPartitionsPerc []int, ramPartitionsPerc []int, minimumNumOfCPU int, minimumAmountOfRAM int) *ResourcesMap {
	rm := &ResourcesMap{}
	rm.numOfCPUCombinations = cap(cpuPartitionsPerc)
	rm.numOfRAMCombinations = cap(ramPartitionsPerc)
	rm.minNumOfCPU = minimumNumOfCPU
	rm.minAmountOfRAM = minimumAmountOfRAM
	rm.resourcesIdMap = make([][]*GuidRange, rm.numOfCPUCombinations)

	cpuBaseGuid := NewGuidInteger(0) // The Guids starts in 0
	cpuPartitions := NewGuidRange(*cpuBaseGuid, *GetMaximumGuid()).CreatePartitions(cpuPartitionsPerc)
	for partIndex, partValue := range cpuPartitions {
		rm.resourcesIdMap[partIndex] = make([]*GuidRange, rm.numOfRAMCombinations) //Allocate the array of ranges for a CPU and RAM capacity
		rm.resourcesIdMap[partIndex] = partValue.CreatePartitions(ramPartitionsPerc)
	}

	return rm
}

func (rm *ResourcesMap) getIndexByCPU(numOfCPU int) int {
	firstPow := int(math.Log2(float64(rm.minNumOfCPU)))
	pow := int(math.Log2(float64(numOfCPU)))
	return pow - firstPow
}

func (rm *ResourcesMap) getIndexByRAM(amountOfRAM int) int {
	firstPow := int(math.Log2(float64(rm.minAmountOfRAM)))
	pow := int(math.Log2(float64(amountOfRAM)))
	return pow - firstPow
}

func (rm *ResourcesMap) getCPUByindex(index int) int {
	firstPow := int(math.Log2(float64(rm.minNumOfCPU)))
	return int(math.Pow(float64(2), float64(firstPow+index)))
}

func (rm *ResourcesMap) getRAMByindex(index int) int {
	firstPow := int(math.Log2(float64(rm.minAmountOfRAM)))
	return int(math.Pow(float64(2), float64(firstPow+index)))
}

func (rm *ResourcesMap) RandomGuid(resources Resources) (*Guid, error) {
	cpuIndex := rm.getIndexByRAM(resources.vCPU)
	ramIndex := rm.getIndexByRAM(resources.RAM)
	return rm.resourcesIdMap[cpuIndex][ramIndex].GenerateRandomBetween()
}

func (rm *ResourcesMap) ResourcesByGuid(rGuid Guid) (*Resources, error) {
	for indexCPU, _ := range rm.resourcesIdMap {
		for indexRAM, _ := range rm.resourcesIdMap {
			if rm.resourcesIdMap[indexCPU][indexRAM].Inside(rGuid) {
				return NewResources(rm.getCPUByindex(indexCPU), rm.getRAMByindex(indexRAM)), nil
			}
		}
	}
	return nil, fmt.Errorf("Invalid GUID: %s", rGuid.ToString())
}

func (rm *ResourcesMap) Print() {
	for indexCPU, _ := range rm.resourcesIdMap {
		fmt.Printf("|%v CPUs:", rm.getCPUByindex(indexCPU))
		for indexRAM, _ := range rm.resourcesIdMap {
			fmt.Printf(" |%vMB RAM: ", rm.getRAMByindex(indexRAM))
			rm.resourcesIdMap[indexCPU][indexRAM].Print()
		}
		fmt.Println("")
	}
}
