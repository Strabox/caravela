package api

import "github.com/strabox/caravela/node/common/guid"

/*
Interface of discovery module for the scheduler
*/
type DiscoveryLocal interface {
	Start()                         // Starts the discovery module operations
	AddTrader(traderGUID guid.Guid) // Add a new trader (called during overlay bootstrap)
}
