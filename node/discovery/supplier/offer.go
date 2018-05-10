package supplier

import (
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/discovery/common"
	"time"
)

/*
Offer that the supplier is advertising into the system.
*/
type supplierOffer struct {
	*common.Offer                // Offer's resources content
	traderResponsible *guid.Guid // Trader's GUID responsible for managing the offer

	lastTimeRefreshed time.Time // Last time the responsible trader has refreshed the offer
	refreshesMissed   int       // Number of times the responsible trader did not send a refresh
}

func newSupplierOffer(id common.OfferID, amount int, res resources.Resources,
	traderResponsible guid.Guid) *supplierOffer {

	offerRes := &supplierOffer{}
	offerRes.Offer = common.NewOffer(id, amount, res)
	offerRes.traderResponsible = &traderResponsible

	offerRes.lastTimeRefreshed = time.Now()
	offerRes.refreshesMissed = 0
	return offerRes
}

func (offer *supplierOffer) IsResponsibleTrader(traderGUID guid.Guid) bool {
	return offer.traderResponsible.Equals(traderGUID)
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
