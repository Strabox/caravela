package api

/*
All the APIs exposed by the CARAVELA node to the outside (other nodes and user)
*/
type Node interface {
	Discovery() Discovery
	Scheduler() Scheduler
}

/*
Interface of discovery module for other CARAVELA's nodes (exposed via the REST API)
*/
type Discovery interface {
	CreateOffer(fromSupplierGUID string, fromSupplierIP string, toTraderGUID string, id int64, amount int, cpus int, ram int)
	RefreshOffer(offerID int64, fromTraderGUID string) bool
	RemoveOffer(fromSupplierIP string, fromSupplierGUID string, toTraderGUID string, offerID int64)
	GetOffers(toTraderGUID string) []Offer
}

/*
Interface of the scheduler module for other CARAVELA's nodes (exposed via the REST API)
*/
type Scheduler interface {
	// User<->Node
	Deploy(containerImageKey string, containerArgs []string, cpus int, ram int)
	// Node<->Node
	Launch(fromBuyerIP string, offerId int64, containerImageKey string, containerArgs []string, cpus int, ram int)
}

/*
Interface exposed to the overlay underneath
*/
type OverlayMembership interface {
	AddTrader(guid []byte)
}

type Offer struct {
	ID         int64
	SupplierIP string
}
