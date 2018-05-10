package api

import (
	"github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
)

/*
Interface of discovery module for the scheduler and containers manager
*/
type DiscoveryInternal interface {
	Start()                                                  // Starts the discovery module operations
	AddTrader(traderGUID guid.Guid)                          // Add a new trader (called during overlay bootstrap)
	Find(resources resources.Resources) []*common.RemoteNode // Find remote nodes that offer the given resources
	ObtainResourcesSlot(offerID int64) *resources.Resources  // Obtain a resource offer to submit a container
	ReturnResourcesSlot(resources resources.Resources)       // Release a resource offer into the system
}
