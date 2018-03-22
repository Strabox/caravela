package discovery

import (
	"github.com/strabox/caravela/api/client"
	"github.com/strabox/caravela/node/configuration"
	"github.com/strabox/caravela/node/guid"
	"github.com/strabox/caravela/node/resources"
	"log"
	"time"
)

type offer struct {
	id           int
	amount       int
	cpus         int
	ram          int
	supplierGUID *guid.Guid
	supplierIP   string
}

type offerKey struct {
	id         int
	supplierIP string
}

type trader struct {
	config              *configuration.Configuration // System configuration values
	client              client.Caravela              // Client for the system
	guid                *guid.Guid                   // Trader's own GUID
	handledResources    *resources.Resources         // Combination of resources that its responsible for manage offer
	refreshOffersTicker *time.Ticker

	offersMap     map[offerKey]*offer
	offersChannel chan *offer
}

func newTrader(config *configuration.Configuration, client client.Caravela, guid guid.Guid,
	resources resources.Resources) *trader {
	res := &trader{}
	res.config = config
	res.client = client
	res.guid = &guid
	res.handledResources = &resources

	res.offersMap = make(map[offerKey]*offer)
	res.offersChannel = make(chan *offer)
	return res
}

/*
Receives a resource offer from other node of the system
*/
func (trader *trader) receiveOffer(id int, amount int, cpus int, ram int, suppGUID string, suppIP string) {
	log.Printf("[Trader] Offer received %dX (CPUs: %d, RAM: %d) from: %s\n", amount, cpus, ram, suppIP)
	// TODO: Save the offer
}

func (trader *trader) handleOffers() {

}
