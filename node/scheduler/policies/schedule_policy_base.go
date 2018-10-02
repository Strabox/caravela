package policies

import (
	"github.com/strabox/caravela/node/common/resources"
)

// BaseSchedulePolicy is the base for the implemented scheduling policies.
type BaseSchedulePolicy struct {
}

func (b *BaseSchedulePolicy) WeightOffers(availableOffers WeightedOffers, necessaryResources resources.Resources) WeightedOffers {
	suitableOffers := make(WeightedOffers, 0)
	for i, offer := range availableOffers {
		offerResources := resources.NewResourcesCPUClass(int(offer.FreeResources.CPUClass), offer.FreeResources.CPUs, offer.FreeResources.Memory)
		// Skip nodes that don't have sufficient available resources.
		if !offerResources.Contains(necessaryResources) {
			continue
		}

		var (
			nodeCpus    = offer.UsedResources.CPUs + offer.FreeResources.CPUs
			nodeMemory  = offer.UsedResources.Memory + offer.FreeResources.Memory
			cpuScore    = 100
			memoryScore = 100
		)

		if necessaryResources.CPUs() > 0 {
			cpuScore = (offer.UsedResources.CPUs + necessaryResources.CPUs()) * 100 / nodeCpus
		}
		if necessaryResources.Memory() > 0 {
			memoryScore = (offer.UsedResources.Memory + necessaryResources.Memory()) * 100 / nodeMemory
		}

		if cpuScore <= 100 && memoryScore <= 100 {
			availableOffers[i].Weight = cpuScore + memoryScore
			suitableOffers = append(suitableOffers, availableOffers[i])
		}
	}
	return suitableOffers
}
