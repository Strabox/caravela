package discovery

import nodeAPI "github.com/strabox/caravela/node/api"

// Discovery API necessary to forward the REST calls
type Discovery interface {
	CreateOffer(fromSupplierGUID string, fromSupplierIP string, toTraderGUID string,
		id int64, amount int, cpus int, ram int)
	RefreshOffer(offerID int64, fromTraderGUID string) bool
	RemoveOffer(fromSupplierIP string, fromSupplierGUID string, toTraderGUID string, offerID int64)
	GetOffers(toTraderGUID string, relay bool, fromNodeGUID string) []nodeAPI.Offer
	AdvertiseNeighborOffers(toTraderGUID string, fromTraderGUID string, traderOfferingIP string,
		traderOfferingGUID string)
}
