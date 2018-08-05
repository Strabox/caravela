package scheduler

import (
	"context"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
)

type discoveryLocal interface {
	Start()
	AddTrader(traderGUID guid.GUID)
	FindOffers(ctx context.Context, resources resources.Resources) []types.AvailableOffer
	ObtainResources(offerID int64, necessaryResources resources.Resources) bool
	ReturnResources(resources resources.Resources)
}
