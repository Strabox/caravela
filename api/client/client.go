package client

type OfferRefreshResponse struct {
	ToSupplierIP string
	OfferID      int
	Success      bool
}

/*
Client for CARAVELA's nodes trade messages with each other
*/
type Caravela interface {
	// Discovery
	CreateOffer(fromSupplierIP string, fromSupplierGUID string, toTraderIP string, toTraderGUID string, offerID int, amount int,
		cpus int, ram int) *ClientError
	RefreshOffer(toSupplierIP string, fromTraderGUID string, offerID int, responseChan chan<- OfferRefreshResponse)
	// Scheduling
	// TODO
}
