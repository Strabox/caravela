package resources

import (
	"github.com/pkg/errors"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/common/guid"
)

// Mapping ...
type Mapping struct {
	partitions        *ResourcePartitions //
	resourcesGUIDMap  [][][]*guid.Range   // Matrix of GUID ranges for each resource combination
	resourcesRangeMap map[float64]map[float64]map[float64]*guid.Range
}

// NewResourcesMap creates a new resource map given the CPUs and RAM partitions and the respective GUID distributions.
func NewResourcesMap(partitions *ResourcePartitions) *Mapping {
	cpuClassPartitions := make([][][]*guid.Range, len(partitions.cpuClassPartitions))

	cpuClassPercentages := partitions.CPUClassPercentages()
	cpuClassRanges := guid.NewGUIDRange(*guid.NewZero(), *guid.MaximumGUID()).CreatePartitions(cpuClassPercentages)
	for i := range cpuClassPartitions {
		currentCPUCoresPercentages := partitions.cpuClassPartitions[i].CPUCoresPercentages()
		currentCPUCoresRanges := cpuClassRanges[i].CreatePartitions(currentCPUCoresPercentages)
		cpuClassPartitions[i] = make([][]*guid.Range, len(partitions.cpuClassPartitions[i].cpuCoresPartitions))

		for k := range cpuClassPartitions[i] {
			currentCPURamPercentages := partitions.cpuClassPartitions[i].cpuCoresPartitions[k].RAMPercentages()
			cpuClassPartitions[i][k] = currentCPUCoresRanges[k].CreatePartitions(currentCPURamPercentages)
		}
	}

	resourcesRangeMap := make(map[float64]map[float64]map[float64]*guid.Range)
	for i, cpuClassPartition := range partitions.cpuClassPartitions {
		resourcesRangeMap[cpuClassPartition.Value] = make(map[float64]map[float64]*guid.Range)
		for k, cpuCoresPartition := range cpuClassPartition.cpuCoresPartitions {
			resourcesRangeMap[cpuClassPartition.Value][cpuCoresPartition.Value] = make(map[float64]*guid.Range)
			for j, ramPartition := range cpuCoresPartition.ramPartitions {
				resourcesRangeMap[cpuClassPartition.Value][cpuCoresPartition.Value][ramPartition.Value] = cpuClassPartitions[i][k][j]
			}
		}
	}

	return &Mapping{
		partitions:        partitions,
		resourcesGUIDMap:  cpuClassPartitions,
		resourcesRangeMap: resourcesRangeMap,
	}
}

func ObtainConfiguredPartitions(configPartitions configuration.ResourcesPartitions) *ResourcePartitions {
	res := &ResourcePartitions{}
	res.cpuClassPartitions = make([]CPUClassPartition, len(configPartitions.CPUClasses))
	for p, cpuClassPart := range configPartitions.CPUClasses {
		res.cpuClassPartitions[p].Value = float64(cpuClassPart.Value)
		res.cpuClassPartitions[p].Percentage = cpuClassPart.Percentage
		res.cpuClassPartitions[p].cpuCoresPartitions = make([]CPUCoresPartition, len(cpuClassPart.CPUCores))
		for c, cpuCoresPart := range cpuClassPart.CPUCores {
			res.cpuClassPartitions[p].cpuCoresPartitions[c].Value = float64(cpuCoresPart.Value)
			res.cpuClassPartitions[p].cpuCoresPartitions[c].Percentage = cpuCoresPart.Percentage
			res.cpuClassPartitions[p].cpuCoresPartitions[c].ramPartitions = make([]RAMPartition, len(cpuCoresPart.RAMs))
			for r, ramPart := range cpuCoresPart.RAMs {
				res.cpuClassPartitions[p].cpuCoresPartitions[c].ramPartitions[r].Value = float64(ramPart.Value)
				res.cpuClassPartitions[p].cpuCoresPartitions[c].ramPartitions[r].Percentage = ramPart.Percentage
			}
		}
	}
	return res
}

// LowerPartitionsOffer
func (m *Mapping) LowerPartitionsOffer(availableResources Resources) ([]Resources, error) {
	lowerPartitions := make([]Resources, 0)
	fittestAvailableRes, err := m.getFittestResourcesOffer(availableResources)
	if err != nil {
		return nil, err
	}

	currentCPUClassIndex, currentCoresIndex, currentRamIndex := m.indexesByResources(*fittestAvailableRes)

	for cpuClassIndex := currentCPUClassIndex; cpuClassIndex >= 0; cpuClassIndex-- {
		for coresIndex := currentCoresIndex; coresIndex >= 0; coresIndex-- {
			for ramIndex := currentRamIndex; ramIndex >= 0; ramIndex-- {
				currentCPUClass := m.partitions.cpuClassPartitions[cpuClassIndex].Value
				currentCores := m.partitions.cpuClassPartitions[cpuClassIndex].cpuCoresPartitions[coresIndex].Value
				currentRam := m.partitions.cpuClassPartitions[cpuClassIndex].cpuCoresPartitions[coresIndex].ramPartitions[ramIndex].Value
				if currentCPUClass <= float64(fittestAvailableRes.CPUClass()) && currentCores <= float64(fittestAvailableRes.CPUs()) && currentRam <= float64(fittestAvailableRes.RAM()) {
					resources := NewResources(0, 0)
					resources.SetCPUClass(int(m.partitions.cpuClassPartitions[cpuClassIndex].Value))
					resources.SetCPUs(int(m.partitions.cpuClassPartitions[cpuClassIndex].cpuCoresPartitions[coresIndex].Value))
					resources.SetRAM(int(m.partitions.cpuClassPartitions[cpuClassIndex].cpuCoresPartitions[coresIndex].ramPartitions[ramIndex].Value))
					lowerPartitions = append(lowerPartitions, *resources)
				}
			}
			if coresIndex == 0 && (cpuClassIndex-1) >= 0 {
				tmpCoresIndex := len(m.partitions.cpuClassPartitions[cpuClassIndex-1].cpuCoresPartitions) - 1
				currentRamIndex = len(m.partitions.cpuClassPartitions[cpuClassIndex-1].cpuCoresPartitions[tmpCoresIndex].ramPartitions) - 1
			} else if (coresIndex - 1) >= 0 {
				currentRamIndex = len(m.partitions.cpuClassPartitions[cpuClassIndex].cpuCoresPartitions[coresIndex-1].ramPartitions) - 1
			}
		}
		if (cpuClassIndex - 1) >= 0 {
			currentCoresIndex = len(m.partitions.cpuClassPartitions[cpuClassIndex-1].cpuCoresPartitions) - 1
		}
	}
	return lowerPartitions, nil
}

// RandGUIDFittestSearch returns a random GUID in the range of the respective "fittest" target resource combination.
func (m *Mapping) RandGUIDFittestSearch(targetResources Resources) (*guid.GUID, error) {
	fittestRes, err := m.getFittestResourcesSearch(targetResources)
	if err != nil {
		return nil, err
	}
	return m.resourcesRangeMap[float64(fittestRes.CPUClass())][float64(fittestRes.CPUs())][float64(fittestRes.RAM())].GenerateRandom()
}

func (m *Mapping) RandGUIDHighestSearch(targetResources Resources) (*guid.GUID, error) {
	fittestRes, err := m.getHighestResourcesSearch(targetResources)
	if err != nil {
		return nil, err
	}
	return m.resourcesRangeMap[float64(fittestRes.CPUClass())][float64(fittestRes.CPUs())][float64(fittestRes.RAM())].GenerateRandom()
}

// RandGUIDOffer returns a random GUID in the range of the respective "fittest" target resource combination.
func (m *Mapping) RandGUIDOffer(targetResources Resources) (*guid.GUID, error) {
	fittestRes, err := m.getFittestResourcesOffer(targetResources)
	if err != nil {
		return nil, err
	}
	return m.resourcesRangeMap[float64(fittestRes.CPUClass())][float64(fittestRes.CPUs())][float64(fittestRes.RAM())].GenerateRandom()
}

// FirstGUIDOffer returns the first GUID that represents the given resources.
func (m *Mapping) FirstGUIDOffer(resources Resources) (*guid.GUID, error) {
	fittestRes, err := m.getFittestResourcesOffer(resources)
	if err != nil {
		return nil, err
	}
	return m.resourcesRangeMap[float64(fittestRes.CPUClass())][float64(fittestRes.CPUs())][float64(fittestRes.RAM())].LowerGUID(), nil
}

// ResourcesByGUID returns the resources combination that maps to the given GUID.
func (m *Mapping) ResourcesByGUID(resGUID guid.GUID) *Resources {
	for indexCPUClass := range m.resourcesGUIDMap {
		for indexCPUCores := range m.resourcesGUIDMap[indexCPUClass] {
			for indexRAM := range m.resourcesGUIDMap[indexCPUClass][indexCPUCores] {
				if m.resourcesGUIDMap[indexCPUClass][indexCPUCores][indexRAM].Inside(resGUID) {
					return m.resourcesByIndexes(indexCPUClass, indexCPUCores, indexRAM)
				}
			}
		}
	}
	return nil
}

// LowestResources returns the lowest resource combination available.
func (m *Mapping) LowestResources() *Resources {
	lowestResources := NewResourcesCPUClass(int(types.LowCPUPClass), 0, 0)
	lowestResources.SetCPUClass(int(m.partitions.cpuClassPartitions[0].Value))
	lowestResources.SetCPUs(int(m.partitions.cpuClassPartitions[0].cpuCoresPartitions[0].Value))
	lowestResources.SetRAM(int(m.partitions.cpuClassPartitions[0].cpuCoresPartitions[0].ramPartitions[0].Value))
	return lowestResources
}

// HigherRandGUIDSearch returns a random GUID in the next range of resources.
// First it tries the GUIDs that represent the SAME cpus and MORE ram.
// Second it tries the GUIDs that represent the MORE cpus and SAME ram.
// Lastly it will try the GUIDs that represent the MORE cpus and MORE ram.
func (m *Mapping) HigherRandGUIDSearch(currentGUID guid.GUID, targetResources Resources) (*guid.GUID, error) {
	currentGuidResources := m.ResourcesByGUID(currentGUID)

	currentCpuClassIndex, currentCoresIndex, currentRamIndex := m.indexesByResources(*currentGuidResources)

	firstHit := true
	for cpuClassIndex := currentCpuClassIndex; cpuClassIndex < len(m.partitions.cpuClassPartitions); cpuClassIndex++ {
		for coresIndex := currentCoresIndex; coresIndex < len(m.partitions.cpuClassPartitions[cpuClassIndex].cpuCoresPartitions); coresIndex++ {
			for ramIndex := currentRamIndex; ramIndex < len(m.partitions.cpuClassPartitions[cpuClassIndex].cpuCoresPartitions[coresIndex].ramPartitions); ramIndex++ {
				currentCPUClass := m.partitions.cpuClassPartitions[cpuClassIndex].Value
				currentCores := m.partitions.cpuClassPartitions[cpuClassIndex].cpuCoresPartitions[coresIndex].Value
				currentRam := m.partitions.cpuClassPartitions[cpuClassIndex].cpuCoresPartitions[coresIndex].ramPartitions[ramIndex].Value
				if currentCPUClass >= float64(targetResources.CPUClass()) && currentCores >= float64(targetResources.CPUs()) && currentRam >= float64(targetResources.RAM()) {
					if firstHit {
						firstHit = false
						continue
					}
					return m.resourcesGUIDMap[cpuClassIndex][coresIndex][ramIndex].GenerateRandom()
				}
			}
			currentRamIndex = 0
		}
		currentCoresIndex = 0
	}

	return nil, errors.New("No more resources combinations")
}

func (m *Mapping) LowerRandGUIDSearch(currentGUID guid.GUID, targetResources Resources) (*guid.GUID, error) {
	currentGuidResources := m.ResourcesByGUID(currentGUID)

	currentCPUClassIndex, currentCoresIndex, currentRamIndex := m.indexesByResources(*currentGuidResources)

	firstHit := true
	for cpuClassIndex := currentCPUClassIndex; cpuClassIndex >= 0; cpuClassIndex-- {
		for coresIndex := currentCoresIndex; coresIndex >= 0; coresIndex-- {
			for ramIndex := currentRamIndex; ramIndex >= 0; ramIndex-- {
				currentCPUClass := m.partitions.cpuClassPartitions[cpuClassIndex].Value
				currentCores := m.partitions.cpuClassPartitions[cpuClassIndex].cpuCoresPartitions[coresIndex].Value
				currentRam := m.partitions.cpuClassPartitions[cpuClassIndex].cpuCoresPartitions[coresIndex].ramPartitions[ramIndex].Value
				if currentCPUClass >= float64(targetResources.CPUClass()) && currentCores >= float64(targetResources.CPUs()) && currentRam >= float64(targetResources.RAM()) {
					if firstHit {
						firstHit = false
						continue
					}
					return m.resourcesGUIDMap[cpuClassIndex][coresIndex][ramIndex].GenerateRandom()
				}
			}
			if coresIndex == 0 && (cpuClassIndex-1) >= 0 {
				tmpCoresIndex := len(m.partitions.cpuClassPartitions[cpuClassIndex-1].cpuCoresPartitions) - 1
				currentRamIndex = len(m.partitions.cpuClassPartitions[cpuClassIndex-1].cpuCoresPartitions[tmpCoresIndex].ramPartitions) - 1
			} else if (coresIndex - 1) >= 0 {
				currentRamIndex = len(m.partitions.cpuClassPartitions[cpuClassIndex].cpuCoresPartitions[coresIndex-1].ramPartitions) - 1
			}
		}
		if cpuClassIndex-1 >= 0 {
			currentCoresIndex = len(m.partitions.cpuClassPartitions[cpuClassIndex-1].cpuCoresPartitions) - 1
		}
	}
	return nil, errors.New("No more resources combinations")
}

// LowerRandGUIDOffer returns a random GUID in the previous range of resources.
// First it tries the GUIDs that represent the SAME cpus and LESS ram.
// Second it tries the GUIDs that represent the LESS cpus and SAME ram.
// Lastly it will try the GUIDs that represent the LESS cpus and LESS ram.
func (m *Mapping) LowerRandGUIDOffer(currentGUID guid.GUID, targetResources Resources) (*guid.GUID, error) {
	currentGuidResources := m.ResourcesByGUID(currentGUID)

	currentCPUClassIndex, currentCoresIndex, currentRamIndex := m.indexesByResources(*currentGuidResources)

	firstHit := true
	for cpuClassIndex := currentCPUClassIndex; cpuClassIndex >= 0; cpuClassIndex-- {
		for coresIndex := currentCoresIndex; coresIndex >= 0; coresIndex-- {
			for ramIndex := currentRamIndex; ramIndex >= 0; ramIndex-- {
				currentCPUClass := m.partitions.cpuClassPartitions[cpuClassIndex].Value
				currentCores := m.partitions.cpuClassPartitions[cpuClassIndex].cpuCoresPartitions[coresIndex].Value
				currentRam := m.partitions.cpuClassPartitions[cpuClassIndex].cpuCoresPartitions[coresIndex].ramPartitions[ramIndex].Value
				if currentCPUClass <= float64(targetResources.CPUClass()) && currentCores <= float64(targetResources.CPUs()) && currentRam <= float64(targetResources.RAM()) {
					if firstHit {
						firstHit = false
						continue
					}
					return m.resourcesGUIDMap[cpuClassIndex][coresIndex][ramIndex].GenerateRandom()
				}
			}
			if coresIndex == 0 && (cpuClassIndex-1) >= 0 {
				tmpCoresIndex := len(m.partitions.cpuClassPartitions[cpuClassIndex-1].cpuCoresPartitions) - 1
				currentRamIndex = len(m.partitions.cpuClassPartitions[cpuClassIndex-1].cpuCoresPartitions[tmpCoresIndex].ramPartitions) - 1
			} else if (coresIndex - 1) >= 0 {
				currentRamIndex = len(m.partitions.cpuClassPartitions[cpuClassIndex].cpuCoresPartitions[coresIndex-1].ramPartitions) - 1
			}
		}
		if cpuClassIndex-1 >= 0 {
			currentCoresIndex = len(m.partitions.cpuClassPartitions[cpuClassIndex-1].cpuCoresPartitions) - 1
		}
	}
	return nil, errors.New("No more resources combinations")
}

//
func (m *Mapping) resourcesByIndexes(cpuClassIndex, cpuCoresIndex, ramIndex int) *Resources {
	cpuClass := int(m.partitions.cpuClassPartitions[cpuClassIndex].Value)
	cpuCores := int(m.partitions.cpuClassPartitions[cpuClassIndex].cpuCoresPartitions[cpuCoresIndex].Value)
	ram := int(m.partitions.cpuClassPartitions[cpuClassIndex].cpuCoresPartitions[cpuCoresIndex].ramPartitions[ramIndex].Value)
	return NewResourcesCPUClass(cpuClass, cpuCores, ram)
}

func (m *Mapping) indexesByResources(fittestResources Resources) (int, int, int) {
	currentCPUClassIndex, currentCoresIndex, currentRamIndex := 0, 0, 0
ExitLoop:
	for cpuClassIndex, cpuClassPartition := range m.partitions.cpuClassPartitions {
		if cpuClassPartition.Value == float64(fittestResources.CPUClass()) {
			for coresIndex, coresPartition := range cpuClassPartition.cpuCoresPartitions {
				if coresPartition.Value == float64(fittestResources.CPUs()) {
					for ramIndex, ramPartition := range coresPartition.ramPartitions {
						if ramPartition.Value == float64(fittestResources.RAM()) {
							currentCPUClassIndex = cpuClassIndex
							currentCoresIndex = coresIndex
							currentRamIndex = ramIndex
							break ExitLoop
						}
					}
				}
			}
		}
	}
	return currentCPUClassIndex, currentCoresIndex, currentRamIndex
}

func (m *Mapping) SamePartitionResourcesSearch(arg1 Resources, arg2 Resources) (bool, error) {
	arg1PartitionRes, err := m.getFittestResourcesSearch(arg1)
	if err != nil {
		return false, err
	}
	arg2PartitionRes, err := m.getFittestResourcesSearch(arg2)
	if err != nil {
		return false, err
	}
	return arg1PartitionRes.Equals(*arg2PartitionRes), nil
}

// getFittestResourcesSearch returns the fittest resources combination that contains the resources given..
// Return the lowest resource section that contains the targetResources.
func (m *Mapping) getFittestResourcesSearch(targetResources Resources) (*Resources, error) {
	fittestRes := NewResourcesCPUClass(0, 0, 0)

ExitLoop:
	for _, cpuClassPartition := range m.partitions.cpuClassPartitions {
		if targetResources.CPUClass() <= int(cpuClassPartition.Value) {
			for _, coresPartition := range cpuClassPartition.cpuCoresPartitions {
				if targetResources.CPUs() <= int(coresPartition.Value) {
					for _, ramPartition := range coresPartition.ramPartitions {
						if targetResources.RAM() <= int(ramPartition.Value) {
							fittestRes.SetCPUClass(int(cpuClassPartition.Value))
							fittestRes.SetCPUs(int(coresPartition.Value))
							fittestRes.SetRAM(int(ramPartition.Value))
							break ExitLoop
						}
					}
				}
			}
		}
	}

	if fittestRes.IsZero() {
		return nil, errors.New("no target targetResources available")
	}
	return fittestRes, nil
}

// Return the highest resource section that contains the targetResources.
func (m *Mapping) getHighestResourcesSearch(targetResources Resources) (*Resources, error) {
	highestRes := NewResourcesCPUClass(0, 0, 0)

ExitLoop:
	for cpuClassIndex := len(m.partitions.cpuClassPartitions) - 1; cpuClassIndex >= 0; cpuClassIndex-- {
		if targetResources.CPUClass() <= int(m.partitions.cpuClassPartitions[cpuClassIndex].Value) {
			cpuCoresPartitions := m.partitions.cpuClassPartitions[cpuClassIndex].cpuCoresPartitions
			for coresIndex := len(cpuCoresPartitions) - 1; coresIndex >= 0; coresIndex-- {
				if targetResources.CPUs() <= int(cpuCoresPartitions[coresIndex].Value) {
					ramPartitions := cpuCoresPartitions[coresIndex].ramPartitions
					for ramIndex := len(ramPartitions) - 1; ramIndex >= 0; ramIndex-- {
						if targetResources.RAM() <= int(ramPartitions[ramIndex].Value) {
							highestRes.SetCPUClass(int(m.partitions.cpuClassPartitions[cpuClassIndex].Value))
							highestRes.SetCPUs(int(cpuCoresPartitions[coresIndex].Value))
							highestRes.SetRAM(int(ramPartitions[ramIndex].Value))
							break ExitLoop
						}
					}
				}
			}
		}
	}

	if highestRes.IsZero() {
		return nil, errors.New("no target targetResources available")
	}
	return highestRes, nil
}

// getFittestResourcesOffer returns the fittest resources combination that can be responsible by the resources.
// Return the highest resource section that is contained by the offerResources.
func (m *Mapping) getFittestResourcesOffer(offerResources Resources) (*Resources, error) {
	fittestRes := NewResourcesCPUClass(0, 0, 0)

ExitLoop:
	for cpuClassIndex := len(m.partitions.cpuClassPartitions) - 1; cpuClassIndex >= 0; cpuClassIndex-- {
		if offerResources.CPUClass() >= int(m.partitions.cpuClassPartitions[cpuClassIndex].Value) {
			cpuCoresPartitions := m.partitions.cpuClassPartitions[cpuClassIndex].cpuCoresPartitions
			for coresIndex := len(cpuCoresPartitions) - 1; coresIndex >= 0; coresIndex-- {
				if offerResources.CPUs() >= int(cpuCoresPartitions[coresIndex].Value) {
					ramPartitions := cpuCoresPartitions[coresIndex].ramPartitions
					for ramIndex := len(ramPartitions) - 1; ramIndex >= 0; ramIndex-- {
						if offerResources.RAM() >= int(ramPartitions[ramIndex].Value) {
							fittestRes.SetCPUClass(int(m.partitions.cpuClassPartitions[cpuClassIndex].Value))
							fittestRes.SetCPUs(int(cpuCoresPartitions[coresIndex].Value))
							fittestRes.SetRAM(int(ramPartitions[ramIndex].Value))
							break ExitLoop
						}
					}
				}
			}
		}
	}
	if fittestRes.IsZero() {
		return nil, errors.New("no target offerResources that can handle available")
	}
	return fittestRes, nil
}
