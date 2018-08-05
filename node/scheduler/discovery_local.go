package scheduler

import (
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
)

type discoverylocal interface {
	Start()
	AddTrader(traderGUID guid.GUID)
	FindOffers(resources resources.Resources) []types.AvailableOffer
	ObtainResources(offerID int64, necessaryResources resources.Resources) bool
	ReturnResources(resources resources.Resources)
}
