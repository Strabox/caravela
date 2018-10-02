package binpack

import (
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/scheduler/policies"
	"sort"
)

// SchedulePolicy implements the SchedulePolicy interface.
// This policy tries to maximize the use of a node first before trying other nodes. More consolidation less load balancing.
type SchedulePolicy struct {
	policies.BaseSchedulePolicy
}

// NewBinPackSchedulePolicy creates a new binpack schedule policy.
func NewBinPackSchedulePolicy() (policies.SchedulingPolicy, error) {
	return &SchedulePolicy{}, nil
}

func (s *SchedulePolicy) Rank(availableOffers policies.WeightedOffers, necessaryResources resources.Resources) policies.WeightedOffers {
	suitableOffers := s.WeightOffers(availableOffers, necessaryResources)
	sort.Sort(sort.Reverse(suitableOffers))
	return suitableOffers
}
