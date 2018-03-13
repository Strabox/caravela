package discovery

import (
	"github.com/strabox/caravela/node/guid"
	"github.com/strabox/caravela/node/resources"
)

type Trader struct {
	guid      *guid.Guid           // Trader's Guid in the overlay
	resources *resources.Resources // Combination of resources that its responsible for manage offer
}

func NewTrader(guid guid.Guid, resources resources.Resources) *Trader {
	res := &Trader{}
	res.guid = &guid
	res.resources = &resources
	return res
}
