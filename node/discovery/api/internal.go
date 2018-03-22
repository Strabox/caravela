package api

import "github.com/strabox/caravela/node/guid"

/*
Interface of discovery module for the scheduler
*/
type DiscoveryLocal interface {
	Start()                         // Starts the discovery module operations
	Find()                          // TODO: Request to find a node for a deployment based on resources
	Deploy()                        // TODO: Request to deploy a container in this node
	AddTrader(traderGUID guid.Guid) // Add a new trader (called during overlay bootstrap)
}
