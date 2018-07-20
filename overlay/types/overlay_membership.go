package types

//Node interface exposed to the overlay below.
type OverlayMembership interface {
	// Called when a new local virtual node joins the overlay.
	AddTrader(guid []byte)
}
