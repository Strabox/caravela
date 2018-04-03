package supplier

import (
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/discovery/common"
	"time"
)

type supplierOffer struct {
	offerContent      *common.Offer // Offer's resources content
	traderResponsible *guid.Guid    // Trader's GUID responsible for managing the offer

	lastTimeRefreshed time.Time // Last time the responsible trader has refreshed the offer
	refreshesMissed   int       // Number of times the responsible trader did not send a refresh
}

func newSupplierOffer(offerContent common.Offer, traderResponsible guid.Guid) *supplierOffer {
	offerRes := &supplierOffer{}
	offerRes.offerContent = &offerContent
	offerRes.traderResponsible = &traderResponsible

	offerRes.lastTimeRefreshed = time.Now()
	offerRes.refreshesMissed = 0
	return offerRes
}

func (offer *supplierOffer) LocalID() common.OfferID {
	return offer.offerContent.ID()
}

func (offer *supplierOffer) Amount() int {
	return offer.offerContent.Amount()
}

func (offer *supplierOffer) Resources() *resources.Resources {
	return offer.offerContent.Resources().Copy()
}

func (offer *supplierOffer) IsResponsibleTrader(traderGUID guid.Guid) bool {
	if offer.traderResponsible.Equals(traderGUID) {
		return true
	} else {
		return false
	}
}

func (offer *supplierOffer) RefreshesMissed() int {
	return offer.refreshesMissed
}

func (offer *supplierOffer) Refresh() {
	offer.lastTimeRefreshed = time.Now()
	offer.refreshesMissed = 0
}

func (offer *supplierOffer) VerifyRefreshMiss(refreshTimeout time.Duration) {
	if time.Now().After(offer.lastTimeRefreshed.Add(refreshTimeout)) {
		offer.refreshesMissed++
	}
}
