package client

/*
Client for CARAVELA's nodes trade messages with each other
*/
type Caravela interface {
	// Discovery
	Offer(destTraderIP string, destTraderGUID string, suppIP string, suppGUID string, offerID int, amount int) *ClientError
	RefreshOffer(destSupplierIP string, traderGUID string, offerID int) *ClientError
	RemoveOffer(destTraderIP string, destTraderGUID string, suppGUID string, offerID int) *ClientError
	// Scheduling
	// TODO
}
