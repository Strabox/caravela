package scheduler

import (
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/api/remote"
	"github.com/strabox/caravela/node/common/resources"
	apiInternal "github.com/strabox/caravela/node/discovery/api"
)

type Scheduler struct {
	discovery apiInternal.DiscoveryInternal
	client    remote.Caravela
}

func NewScheduler(internalDisc apiInternal.DiscoveryInternal, client remote.Caravela) *Scheduler {
	res := &Scheduler{}
	res.discovery = internalDisc
	res.client = client
	return res
}

func (s *Scheduler) Deploy(containerImageKey string, containerArgs []string, cpus int, ram int) {
	log.Debugf("[Scheduler:Deploy] %s , CPUs: %d, RAM: %d", containerImageKey, cpus, ram)
	remoteNodes := s.discovery.Find(*resources.NewResources(cpus, ram))
	log.Debugln("[Scheduler:Deploy] Remote Nodes: ", remoteNodes)
	for _, v := range remoteNodes {
		log.Debugln("[Scheduler:Deploy] Node: ", v.GUID().String())
		_, offers := s.client.GetOffers(v.IPAddress(), v.GUID().String())
		log.Debugln("[Scheduler:Deploy] Offers: ", offers)
	}
	// TODO
}

func (s *Scheduler) Launch(containerImageKey string, containerArgs []string, cpus int, ram int, offerID int64) {
	// TODO
}

func (s *Scheduler) LaunchNotification() {
	// TODO
}
