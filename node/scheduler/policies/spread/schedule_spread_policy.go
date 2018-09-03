package spread

import (
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/scheduler/policies"
	"sort"
)

// SchedulePolicy implements the SchedulePolicy interface.
// This policy promotes the spread of container by all nodes. More load balancing less consolidation.
type SchedulePolicy struct {
	policies.BaseSchedulePolicy
}

// NewSpreadSchedulePolicy creates a new spread schedule policy.
func NewSpreadSchedulePolicy() (policies.SchedulingPolicy, error) {
	return &SchedulePolicy{}, nil
}

func (s *SchedulePolicy) Sort(availableOffers policies.WeightedOffers, necessaryResources resources.Resources) {
	s.Rank(availableOffers, necessaryResources)
	sort.Sort(availableOffers)
}
