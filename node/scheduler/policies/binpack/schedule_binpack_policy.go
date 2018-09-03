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

func (s *SchedulePolicy) Sort(availableOffers policies.WeightedOffers, necessaryResources resources.Resources) {
	s.Rank(availableOffers, necessaryResources)
	sort.Sort(sort.Reverse(availableOffers))
}
