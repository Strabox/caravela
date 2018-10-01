package overlay

// LocalNode exposes the overlay to the application node.
type LocalNode interface {
	// AddTrader is called when a new local virtual node joins the overlay.
	AddTrader(guid []byte)
	// GUID returns the node's GUID.
	GUID() string
}
