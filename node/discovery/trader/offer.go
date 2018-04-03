package trader

import (
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/discovery/common"
	"time"
)

/*
offerKey is based on the local offer id and the supplier IP TODO: Hoping the supplier IP is unique????
*/
type offerKey struct {
	id         common.OfferID
	supplierIP string
}

type traderOffer struct {
	supplierGUID *guid.Guid    // GUID of the supplier offering these resources
	supplierIP   string        // IP of the supplier offering these resources
	offer        *common.Offer // Offer resources

	lastRefreshTimestamp time.Time // Last time the offer was refreshed with/without success
	waitingForRefresh    bool      // Marks if there is still a refresh pending for the offer (avoid multiple refreshes)
	refreshesFailed      int       // Number of times the supplier didn't answer to the refresh message
}

func newTraderOffer(supplierGUID guid.Guid, supplierIP string, offer common.Offer) *traderOffer {
	offerRes := &traderOffer{}
	offerRes.supplierGUID = &supplierGUID
	offerRes.supplierIP = supplierIP
	offerRes.offer = &offer

	offerRes.lastRefreshTimestamp = time.Now()
	offerRes.waitingForRefresh = false
	offerRes.refreshesFailed = 0
	return offerRes
}

func (offer *traderOffer) SupplierIP() string {
	return offer.supplierIP
}

func (offer *traderOffer) LocalID() common.OfferID {
	return offer.offer.ID()
}

func (offer *traderOffer) Amount() int {
	return offer.offer.Amount()
}

func (offer *traderOffer) Resources() *resources.Resources {
	return offer.offer.Resources().Copy()
}

func (offer *traderOffer) RefreshesFailed() int {
	return offer.refreshesFailed
}

func (offer *traderOffer) Refresh() bool {
	if offer.waitingForRefresh {
		return false
	} else {
		offer.waitingForRefresh = true
		return true
	}
}

func (offer *traderOffer) RefreshFailed() {
	offer.refreshesFailed++
	offer.lastRefreshTimestamp = time.Now()
	offer.waitingForRefresh = false
}

func (offer *traderOffer) RefreshSucceeded() {
	offer.refreshesFailed = 0
	offer.lastRefreshTimestamp = time.Now()
	offer.waitingForRefresh = false
}
