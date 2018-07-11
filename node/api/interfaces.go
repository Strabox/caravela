package api

//All the APIs exposed by the CARAVELA node to the outside (other nodes and user)
type Node interface {
	Start(join bool, joinIP string) error
	Stop()
}

//Node interface exposed to the overlay below.
type OverlayMembership interface {
	// Called when a new local virtual node joins the overlay.
	AddTrader(guid []byte)
}

type Offer struct {
	ID         int64
	SupplierIP string
}
