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
	*common.Offer // Offer's resources content

	responsibleTraderGUID *guid.GUID // Trader's GUID responsible for managing the offer
	responsibleTraderIP   string     // Trader's IP responsible for managing the offer

	lastTimeRefreshed time.Time // Last time the responsible trader has refreshed the offer
	refreshesMissed   int       // Number of times the responsible trader did not send a refresh
}

func newSupplierOffer(id common.OfferID, amount int, res resources.Resources,
	responsibleTraderIP string, responsibleTraderGUID guid.GUID) *supplierOffer {

	offerRes := &supplierOffer{}
	offerRes.Offer = common.NewOffer(id, amount, res)

	offerRes.responsibleTraderGUID = &responsibleTraderGUID
	offerRes.responsibleTraderIP = responsibleTraderIP

	offerRes.lastTimeRefreshed = time.Now()
	offerRes.refreshesMissed = 0
	return offerRes
}

/*
Verify if the GUID given is from the trader responsible for the offer.
*/
func (offer *supplierOffer) IsResponsibleTrader(traderGUID guid.GUID) bool {
	return offer.responsibleTraderGUID.Equals(traderGUID)
}

/*
Refresh the offer. Called when the supplier received a refresh message for this offer from
the responsible trader.
*/
func (offer *supplierOffer) Refresh() {
	offer.lastTimeRefreshed = time.Now()
	offer.refreshesMissed = 0
}

/*
Verify if the a refresh missed given a specific timeout.
*/
func (offer *supplierOffer) VerifyRefreshes(refreshTimeout time.Duration) {
	if time.Now().After(offer.lastTimeRefreshed.Add(refreshTimeout)) {
		offer.refreshesMissed++
	}
}

func (offer *supplierOffer) RefreshesMissed() int {
	return offer.refreshesMissed
}

func (offer *supplierOffer) ResponsibleTraderGUID() *guid.GUID {
	return offer.responsibleTraderGUID.Copy()
}

func (offer *supplierOffer) ResponsibleTraderIP() string {
	return offer.responsibleTraderIP
}
