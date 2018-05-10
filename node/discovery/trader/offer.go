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
	*common.Offer // Offer resources

	supplierGUID *guid.Guid // GUID of the supplier offering these resources
	supplierIP   string     // IP of the supplier offering these resources

	lastRefreshTimestamp time.Time // Last time the offer was refreshed with/without success
	waitingForRefresh    bool      // Marks if there is still a refresh pending for the offer (avoid multiple refreshes)
	refreshesFailed      int       // Number of times the supplier didn't answer to the refresh message
}

func newTraderOffer(supplierGUID guid.Guid, supplierIP string, id common.OfferID, amount int,
	res resources.Resources) *traderOffer {

	offerRes := &traderOffer{}
	offerRes.Offer = common.NewOffer(id, amount, res)

	offerRes.supplierGUID = &supplierGUID
	offerRes.supplierIP = supplierIP

	offerRes.lastRefreshTimestamp = time.Now()
	offerRes.waitingForRefresh = false
	offerRes.refreshesFailed = 0
	return offerRes
}

func (offer *traderOffer) SupplierIP() string {
	return offer.supplierIP
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
