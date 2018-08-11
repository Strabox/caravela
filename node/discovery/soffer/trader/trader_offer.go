package trader

import (
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/discovery/common"
	"time"
)

// offerKey is based on the local offer id and the supplier IP
// TODO: Hoping the supplier IP is unique????
type offerKey struct {
	supplierIP string         // Offer supplier's IP address
	id         common.OfferID // Local id of the offer (in supplier)
}

// Represents an offer from a supplier that the trader is responsible for managing
type traderOffer struct {
	*common.Offer // Offer resources

	supplierGUID *guid.GUID // GUID of the supplier offering these resources
	supplierIP   string     // IP of the supplier offering these resources

	lastRefreshTimestamp time.Time // Last time the offer was refreshed with/without success
	waitingForRefresh    bool      // Marks if there is still a refresh pending for the offer (avoid multiple refreshes)
	refreshesFailed      int       // Number of times the supplier didn't answer to the refresh message
}

func newTraderOffer(supplierGUID guid.GUID, supplierIP string, id common.OfferID, amount int,
	res resources.Resources) *traderOffer {

	return &traderOffer{
		Offer: common.NewOffer(id, amount, res),

		supplierGUID: &supplierGUID,
		supplierIP:   supplierIP,

		lastRefreshTimestamp: time.Now(),
		waitingForRefresh:    false,
		refreshesFailed:      0,
	}
}

// Return true if it is time to refresh the offer, and false otherwise.
func (offer *traderOffer) Refresh() bool {
	if offer.waitingForRefresh {
		return false
	} else {
		offer.waitingForRefresh = true
		return true
	}
}

// Refresh for the offer failed. Supplier didn't replied to refresh, for example.
func (offer *traderOffer) RefreshFailed() {
	offer.refreshesFailed++
	offer.lastRefreshTimestamp = time.Now()
	offer.waitingForRefresh = false
}

// Refresh for the offer succeeded.
func (offer *traderOffer) RefreshSucceeded() {
	offer.refreshesFailed = 0
	offer.lastRefreshTimestamp = time.Now()
	offer.waitingForRefresh = false
}

func (offer *traderOffer) SupplierIP() string {
	return offer.supplierIP
}

func (offer *traderOffer) RefreshesFailed() int {
	return offer.refreshesFailed
}
