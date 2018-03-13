package rest

import (
	"github.com/strabox/caravela/node/discovery"
	"github.com/strabox/caravela/node/scheduler"
)

type NodeRemote interface {
	Discovery() discovery.DiscoveryRemote
	Scheduler() scheduler.SchedulerRemote
}
