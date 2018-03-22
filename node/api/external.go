package api

/*
Interface of discovery module for other CARAVELA's nodes (exposed via the API)
*/
type Discovery interface {
	CreateOffer(id int, amount int, supplierGUID string, supplierIP string)
	RefreshOffer(id int, traderGUID string)
	RemoveOffer(id int, destTraderGUID string, supplierIP string)
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
