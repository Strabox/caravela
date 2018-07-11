package common

import "github.com/strabox/caravela/node/common/resources"

// OfferID is a type for the offer identifier.
type OfferID int64

// Represents the basic offer of resources into the system
type Offer struct {
	id        OfferID              // Local id (for supplier) of the offer
	amount    int                  // Amount of times the resource combination we have
	resources *resources.Resources // Resource combinations of the offer
}

func NewOffer(id OfferID, amount int, res resources.Resources) *Offer {
	offer := &Offer{}
	offer.id = id
	offer.amount = amount
	offer.resources = &res
	return offer
}

func (offer *Offer) ID() OfferID {
	return offer.id
}

func (offer *Offer) Amount() int {
	return offer.amount
}

func (offer *Offer) Resources() *resources.Resources {
	return offer.resources.Copy()
}
