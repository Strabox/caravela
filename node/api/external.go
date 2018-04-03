package api

import "github.com/strabox/caravela/docker"

/*
Interface of discovery module for other CARAVELA's nodes (exposed via the REST API)
*/
type Discovery interface {
	CreateOffer(fromSupplierGUID string, fromSupplierIP string, toTraderGUID string, id int64, amount int, cpus int, ram int)
	RefreshOffer(offerID int64, fromTraderGUID string) bool
	RemoveOffer(fromSupplierIP string, fromSupplierGUID string, toTraderGUID string, offerID int64)
}

/*
Interface of the scheduler module for other CARAVELA's nodes (exposed via the API)
*/
type Scheduler interface {
	// TODO
}

/*
All the APIs exposed by the CARAVELA node to other nodes (exposed via the API)
*/
type Node interface {
	Discovery() Discovery
	Scheduler() Scheduler
	Docker() *docker.Client
}

type LocalNode interface {
	AddTrader(guid []byte)
}
