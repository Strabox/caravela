package policies

import (
	"github.com/strabox/caravela/node/common/resources"
)

// SchedulingPolicy is an interface that can be implemented in order to provide different criteria to rank a given set
// of offers in a different way.
type SchedulingPolicy interface {
	// Sort the given availableOffers knowing the necessary resources for the deployment.
	Rank(availableOffers WeightedOffers, necessaryResources resources.Resources) WeightedOffers
}
