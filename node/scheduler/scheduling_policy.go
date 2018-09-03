package scheduler

import (
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/node/common/resources"
)

type SchedulingPolicy interface {
	Sort(availableOffers weightedOffers, necessary resources.Resources)
}

type weightedOffers []types.AvailableOffer

func (ao weightedOffers) Len() int {
	return len(ao)
}

func (ao weightedOffers) Swap(i, j int) {
	ao[i], ao[j] = ao[j], ao[i]
}
func (ao weightedOffers) Less(i, j int) bool {
	var (
		ip = ao[i]
		jp = ao[j]
	)

	// If the nodes have the same weight sort them out by number of containers.
	if ip.Weight == jp.Weight {
		return true
	}
	return ip.Weight < jp.Weight
}
