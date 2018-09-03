package policies

import "github.com/strabox/caravela/api/types"

// WeightedOffers are used to rank a set offers by its weight calculated according to a specific scheduling policy.
type WeightedOffers []types.AvailableOffer

// ============================== Sort Interface ================================

func (ao WeightedOffers) Len() int {
	return len(ao)
}

func (ao WeightedOffers) Swap(i, j int) {
	ao[i], ao[j] = ao[j], ao[i]
}
func (ao WeightedOffers) Less(i, j int) bool {
	var (
		ip = ao[i]
		jp = ao[j]
	)

	// TODO: If the nodes have the same weight sort them out by number of containers.
	if ip.Weight == jp.Weight {
		return true
	}
	return ip.Weight < jp.Weight
}
