package common

import "github.com/strabox/caravela/node/common/resources"

type OfferID int64

type Offer interface {
	ID() OfferID
	Amount() int
	Resources() *resources.Resources
}

type offer struct {
	id        OfferID
	amount    int
	resources *resources.Resources
}

func NewOffer(id OfferID, amount int, res resources.Resources) *offer {
	offer := &offer{}
	offer.id = id
	offer.amount = amount
	offer.resources = &res
	return offer
}

func (offer *offer) ID() OfferID {
	return offer.id
}

func (offer *offer) Amount() int {
	return offer.amount
}

func (offer *offer) Resources() *resources.Resources {
	return offer.resources.Copy()
}
