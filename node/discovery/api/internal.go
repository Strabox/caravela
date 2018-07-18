package api

import (
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
)

// Interface of discovery module for the scheduler and containers manager
type DiscoveryInternal interface {
	Start()                                                                     // Starts the discovery module operations
	AddTrader(traderGUID guid.GUID)                                             // Add a new trader (called during overlay bootstrap)
	FindOffers(resources resources.Resources) []types.AvailableOffer            // Find a list of offers for a given resource combination
	ObtainResources(offerID int64, necessaryResources resources.Resources) bool // Obtain a resource offer to submit a container
	ReturnResources(resources resources.Resources)                              // Release a resource offer into the system
}
