package resources

import (
	"github.com/pkg/errors"
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
	cpuPowerPartitions := make([][][]*guid.Range, len(partitions.cpuPowerPartitions))

	cpuPowerPercentages := partitions.CPUPowerPercentages()
	cpuPowerRanges := guid.NewGUIDRange(*guid.NewZero(), *guid.MaximumGUID()).CreatePartitions(cpuPowerPercentages)
	for i := range cpuPowerPartitions {
		currentCPUCoresPercentages := partitions.cpuPowerPartitions[i].CPUCoresPercentages()
		currentCPUCoresRanges := cpuPowerRanges[i].CreatePartitions(currentCPUCoresPercentages)
		cpuPowerPartitions[i] = make([][]*guid.Range, len(partitions.cpuPowerPartitions[i].cpuCoresPartitions))

		for k := range cpuPowerPartitions[i] {
			currentCPURamPercentages := partitions.cpuPowerPartitions[i].cpuCoresPartitions[k].RAMPercentages()
			cpuPowerPartitions[i][k] = currentCPUCoresRanges[k].CreatePartitions(currentCPURamPercentages)
		}
	}

	resourcesRangeMap := make(map[float64]map[float64]map[float64]*guid.Range)
	for i, powerPartition := range partitions.cpuPowerPartitions {
		resourcesRangeMap[powerPartition.Value] = make(map[float64]map[float64]*guid.Range)
		for k, coresPartition := range powerPartition.cpuCoresPartitions {
			resourcesRangeMap[powerPartition.Value][coresPartition.Value] = make(map[float64]*guid.Range)
			for j, ramPartition := range coresPartition.ramPartitions {
				resourcesRangeMap[powerPartition.Value][coresPartition.Value][ramPartition.Value] = cpuPowerPartitions[i][k][j]
			}
		}
	}

	return &Mapping{
		partitions:        partitions,
		resourcesGUIDMap:  cpuPowerPartitions,
		resourcesRangeMap: resourcesRangeMap,
	}
}

func ObtainConfiguredPartitions(configPartitions configuration.ResourcesPartitions) *ResourcePartitions {
	res := &ResourcePartitions{}
	res.cpuPowerPartitions = make([]CPUPowerPartition, len(configPartitions.CPUPowers))
	for p, powerPart := range configPartitions.CPUPowers {
		res.cpuPowerPartitions[p].Value = float64(powerPart.Value)
		res.cpuPowerPartitions[p].Percentage = powerPart.Percentage
		res.cpuPowerPartitions[p].cpuCoresPartitions = make([]CPUCoresPartition, len(powerPart.CPUCores))
		for c, corePart := range powerPart.CPUCores {
			res.cpuPowerPartitions[p].cpuCoresPartitions[c].Value = float64(corePart.Value)
			res.cpuPowerPartitions[p].cpuCoresPartitions[c].Percentage = corePart.Percentage
			res.cpuPowerPartitions[p].cpuCoresPartitions[c].ramPartitions = make([]RAMPartition, len(corePart.RAMs))
			for r, ramPart := range corePart.RAMs {
				res.cpuPowerPartitions[p].cpuCoresPartitions[c].ramPartitions[r].Value = float64(ramPart.Value)
				res.cpuPowerPartitions[p].cpuCoresPartitions[c].ramPartitions[r].Percentage = ramPart.Percentage
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

	currentCoresIndex, currentRamIndex := 0, 0
ExitLoop:
	for coresIndex, coresPartition := range m.partitions.cpuPowerPartitions[0].cpuCoresPartitions {
		if coresPartition.Value == float64(fittestAvailableRes.CPUs()) {
			for ramIndex, ramPartition := range coresPartition.ramPartitions {
				if ramPartition.Value == float64(fittestAvailableRes.RAM()) {
					currentCoresIndex = coresIndex
					currentRamIndex = ramIndex
					break ExitLoop
				}
			}
		}
	}

	for coresIndex := currentCoresIndex; coresIndex >= 0; coresIndex-- {
		for ramIndex := currentRamIndex; ramIndex >= 0; ramIndex-- {
			currentCores := m.partitions.cpuPowerPartitions[0].cpuCoresPartitions[coresIndex].Value
			currentRam := m.partitions.cpuPowerPartitions[0].cpuCoresPartitions[coresIndex].ramPartitions[ramIndex].Value
			if currentCores <= float64(fittestAvailableRes.CPUs()) && currentRam <= float64(fittestAvailableRes.RAM()) {
				resources := NewResources(0, 0)
				resources.SetCPUs(int(m.partitions.cpuPowerPartitions[0].cpuCoresPartitions[coresIndex].Value))
				resources.SetRAM(int(m.partitions.cpuPowerPartitions[0].cpuCoresPartitions[coresIndex].ramPartitions[ramIndex].Value))
				lowerPartitions = append(lowerPartitions, *resources)
			}
		}
		if coresIndex-1 >= 0 {
			currentRamIndex = len(m.partitions.cpuPowerPartitions[0].cpuCoresPartitions[coresIndex-1].ramPartitions) - 1
		}
	}
	return lowerPartitions, nil
}

// RandGUIDSearch returns a random GUID in the range of the respective "fittest" target resource combination.
func (m *Mapping) RandGUIDSearch(targetResources Resources) (*guid.GUID, error) {
	fittestRes, err := m.getFittestResourcesSearch(targetResources)
	if err != nil {
		return nil, err
	}
	return m.resourcesRangeMap[0][float64(fittestRes.CPUs())][float64(fittestRes.RAM())].GenerateRandom()
}

// RandGUIDOffer returns a random GUID in the range of the respective "fittest" target resource combination.
func (m *Mapping) RandGUIDOffer(targetResources Resources) (*guid.GUID, error) {
	fittestRes, err := m.getFittestResourcesOffer(targetResources)
	if err != nil {
		return nil, err
	}
	return m.resourcesRangeMap[0][float64(fittestRes.CPUs())][float64(fittestRes.RAM())].GenerateRandom()
}

// FirstGUIDOffer returns the first GUID that represents the given resources.
func (m *Mapping) FirstGUIDOffer(resources Resources) (*guid.GUID, error) {
	fittestRes, err := m.getFittestResourcesOffer(resources)
	if err != nil {
		return nil, err
	}
	return m.resourcesRangeMap[0][float64(fittestRes.CPUs())][float64(fittestRes.RAM())].LowerGUID(), nil
}

// ResourcesByGUID returns the resources combination that maps to the given GUID.
func (m *Mapping) ResourcesByGUID(resGUID guid.GUID) *Resources {
	for indexCPUCores := range m.resourcesGUIDMap[0] {
		for indexRAM := range m.resourcesGUIDMap[0][indexCPUCores] {
			if m.resourcesGUIDMap[0][indexCPUCores][indexRAM].Inside(resGUID) {
				return m.resourcesByIndexes(0, indexCPUCores, indexRAM)
			}
		}
	}
	return nil
}

// LowestResources returns the lowest resource combination available.
func (m *Mapping) LowestResources() *Resources {
	lowestResources := NewResources(0, 0)
	// m.partitions.cpuPowerPartitions[0].Value Not used for now
	lowestResources.SetCPUs(int(m.partitions.cpuPowerPartitions[0].cpuCoresPartitions[0].Value))
	lowestResources.SetRAM(int(m.partitions.cpuPowerPartitions[0].cpuCoresPartitions[0].ramPartitions[0].Value))
	return lowestResources
}

// HigherRandGUIDSearch returns a random GUID in the next range of resources.
// First it tries the GUIDs that represent the SAME cpus and MORE ram.
// Second it tries the GUIDs that represent the MORE cpus and SAME ram.
// Lastly it will try the GUIDs that represent the MORE cpus and MORE ram.
func (m *Mapping) HigherRandGUIDSearch(currentGUID guid.GUID, targetResources Resources) (*guid.GUID, error) {
	currentGuidResources := m.ResourcesByGUID(currentGUID)
	currentCoresIndex := 0
	currentRamIndex := 0
ExitLoop:
	for coresIndex, coresPartition := range m.partitions.cpuPowerPartitions[0].cpuCoresPartitions {
		if coresPartition.Value == float64(currentGuidResources.CPUs()) {
			for ramIndex, ramPartition := range coresPartition.ramPartitions {
				if ramPartition.Value == float64(currentGuidResources.RAM()) {
					currentCoresIndex = coresIndex
					currentRamIndex = ramIndex
					break ExitLoop
				}
			}
		}
	}

	firstHit := true
	for coresIndex := currentCoresIndex; coresIndex < len(m.partitions.cpuPowerPartitions[0].cpuCoresPartitions); coresIndex++ {
		for ramIndex := currentRamIndex; ramIndex < len(m.partitions.cpuPowerPartitions[0].cpuCoresPartitions[coresIndex].ramPartitions); ramIndex++ {
			currentCores := m.partitions.cpuPowerPartitions[0].cpuCoresPartitions[coresIndex].Value
			currentRam := m.partitions.cpuPowerPartitions[0].cpuCoresPartitions[coresIndex].ramPartitions[ramIndex].Value
			if currentCores >= float64(targetResources.CPUs()) && currentRam >= float64(targetResources.RAM()) {
				if firstHit {
					firstHit = false
					continue
				}
				return m.resourcesGUIDMap[0][coresIndex][ramIndex].GenerateRandom()
			}
		}
		currentRamIndex = 0
	}

	return nil, errors.New("No more resources combinations")
}

// LowerRandGUIDOffer returns a random GUID in the previous range of resources.
// First it tries the GUIDs that represent the SAME cpus and LESS ram.
// Second it tries the GUIDs that represent the LESS cpus and SAME ram.
// Lastly it will try the GUIDs that represent the LESS cpus and LESS ram.
func (m *Mapping) LowerRandGUIDOffer(currentGUID guid.GUID, targetResources Resources) (*guid.GUID, error) {
	currentGuidResources := m.ResourcesByGUID(currentGUID)

	currentCoresIndex, currentRamIndex := 0, 0
ExitLoop:
	for coresIndex, coresPartition := range m.partitions.cpuPowerPartitions[0].cpuCoresPartitions {
		if coresPartition.Value == float64(currentGuidResources.CPUs()) {
			for ramIndex, ramPartition := range coresPartition.ramPartitions {
				if ramPartition.Value == float64(currentGuidResources.RAM()) {
					currentCoresIndex = coresIndex
					currentRamIndex = ramIndex
					break ExitLoop
				}
			}
		}
	}

	firstHit := true
	for coresIndex := currentCoresIndex; coresIndex >= 0; coresIndex-- {
		for ramIndex := currentRamIndex; ramIndex >= 0; ramIndex-- {
			currentCores := m.partitions.cpuPowerPartitions[0].cpuCoresPartitions[coresIndex].Value
			currentRam := m.partitions.cpuPowerPartitions[0].cpuCoresPartitions[coresIndex].ramPartitions[ramIndex].Value
			if currentCores <= float64(targetResources.CPUs()) && currentRam <= float64(targetResources.RAM()) {
				if firstHit {
					firstHit = false
					continue
				}
				return m.resourcesGUIDMap[0][coresIndex][ramIndex].GenerateRandom()
			}
		}
		if coresIndex-1 >= 0 {
			currentRamIndex = len(m.partitions.cpuPowerPartitions[0].cpuCoresPartitions[coresIndex-1].ramPartitions) - 1
		}
	}

	return nil, errors.New("No more resources combinations")
}

//
func (m *Mapping) resourcesByIndexes(cpuPowerIndex, cpuCoresIndex, ramIndex int) *Resources {
	//cpuPower := int(m.partitions.cpuPowerPartitions[cpuPowerIndex].Value)
	cpuCores := int(m.partitions.cpuPowerPartitions[cpuPowerIndex].cpuCoresPartitions[cpuCoresIndex].Value)
	ram := int(m.partitions.cpuPowerPartitions[cpuPowerIndex].cpuCoresPartitions[cpuCoresIndex].ramPartitions[ramIndex].Value)
	return NewResources(cpuCores, ram)
}

// getFittestResourcesSearch returns the fittest resources combination that contains the resources given..
func (m *Mapping) getFittestResourcesSearch(resources Resources) (*Resources, error) {
	fittestRes := NewResources(0, 0)

	cpuCoresPartitions := m.partitions.cpuPowerPartitions[0].cpuCoresPartitions // Hack: Hardcoded for now!!!
ExitLoop:
	for _, coresPartition := range cpuCoresPartitions {
		if resources.CPUs() <= int(coresPartition.Value) {
			fittestRes.SetCPUs(int(coresPartition.Value))
			for _, ramPartition := range coresPartition.ramPartitions {
				if resources.RAM() <= int(ramPartition.Value) {
					fittestRes.SetRAM(int(ramPartition.Value))
					break ExitLoop
				}
			}
		}
	}
	if fittestRes.IsZero() {
		return nil, errors.New("no target resources available")
	}
	return fittestRes, nil
}

// getFittestResourcesOffer returns the fittest resources combination that can be responsible by the resources.
func (m *Mapping) getFittestResourcesOffer(resources Resources) (*Resources, error) {
	fittestRes := NewResources(0, 0)

	cpuCoresPartitions := m.partitions.cpuPowerPartitions[0].cpuCoresPartitions // Hack: Hardcoded for now!!!
ExitLoop:
	for coresIndex := len(cpuCoresPartitions) - 1; coresIndex >= 0; coresIndex-- {
		if resources.CPUs() >= int(cpuCoresPartitions[coresIndex].Value) {
			ramPartitions := cpuCoresPartitions[coresIndex].ramPartitions
			for ramIndex := len(ramPartitions) - 1; ramIndex >= 0; ramIndex-- {
				if resources.RAM() >= int(ramPartitions[ramIndex].Value) {
					fittestRes.SetCPUs(int(cpuCoresPartitions[coresIndex].Value))
					fittestRes.SetRAM(int(ramPartitions[ramIndex].Value))
					break ExitLoop
				}
			}
		}
	}
	if fittestRes.IsZero() {
		return nil, errors.New("no target resources that can handle available")
	}
	return fittestRes, nil
}
