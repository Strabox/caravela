package api

/*
Interface of discovery module for other CARAVELA's nodes (exposed via the API)
*/
type Discovery interface {
	CreateOffer(fromSupplierGUID string, fromSupplierIP string, toTraderGUID string, id int, amount int, cpus int, ram int)
	RefreshOffer(id int, fromTraderGUID string) bool
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
}

type LocalNode interface {
	AddTrader(guid []byte)
}
