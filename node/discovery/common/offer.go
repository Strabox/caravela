package common

import "github.com/strabox/caravela/node/common/resources"

type OfferID int64

type Offer struct {
	id        OfferID
	amount    int
	resources *resources.Resources
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
