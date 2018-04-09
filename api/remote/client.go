package remote

/*
Client Offer struct.
*/
type Offer struct {
	ID         int64
	SupplierIP string
}

/*
Client for CARAVELA's nodes trade messages with each other
*/
type Caravela interface {
	// =============================== Discovery ===============================
	/* Sends a create offer message to a trader from a supplier that wants to offer its resources. */
	CreateOffer(fromSupplierIP string, fromSupplierGUID string, toTraderIP string, toTraderGUID string, offerID int64, amount int,
		cpus int, ram int) *Error

	/* Sends a refresh message from a trader to a supplier. It is used to mutually know that both are alive. */
	RefreshOffer(toSupplierIP string, fromTraderGUID string, offerID int64) (*Error, bool)

	/* Sends a remove offer message from a supplier to a trader. It means the supplier does not handle the offer anymore. */
	RemoveOffer(fromSupplierIP string, fromSupplierGUID, toTraderIP string, toTraderGUID string, offerID int64) *Error

	/* Sends a get message to obtain all the available offers in a trader. */
	GetOffers(toTraderIP string, toTraderGUID string) (*Error, []Offer)

	// =============================== Scheduling ===============================
	// TODO: Scheduling API
}
