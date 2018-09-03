package scheduler

import (
	"github.com/strabox/caravela/node/common/resources"
	"sort"
)

type SpreadSchedulingPolicy struct {
}

func (s *SpreadSchedulingPolicy) Sort(availableOffers weightedOffers, necessaryResources resources.Resources) {
	for i, offer := range availableOffers {
		offerResources := resources.NewResourcesCPUClass(int(offer.FreeResources.CPUClass), offer.FreeResources.CPUs, offer.FreeResources.RAM)
		// Skip nodes that don't have sufficient available resources.
		if !offerResources.Contains(necessaryResources) {
			continue
		}

		var (
			nodeCpus    = offer.UsedResources.CPUs + offer.FreeResources.CPUs
			nodeMemory  = offer.UsedResources.RAM + offer.FreeResources.RAM
			cpuScore    = 100
			memoryScore = 100
		)

		if necessaryResources.CPUs() > 0 {
			cpuScore = (offer.UsedResources.CPUs + necessaryResources.CPUs()) * 100 / nodeCpus
		}
		if necessaryResources.RAM() > 0 {
			memoryScore = (offer.UsedResources.RAM + necessaryResources.RAM()) * 100 / nodeMemory
		}

		if cpuScore <= 100 && memoryScore <= 100 {
			availableOffers[i].Weight = cpuScore + memoryScore
		}
	}
	sort.Sort(availableOffers)
}
