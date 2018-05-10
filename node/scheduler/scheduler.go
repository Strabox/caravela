package scheduler

import (
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/api/remote"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/containers"
	apiInternal "github.com/strabox/caravela/node/discovery/api"
	"github.com/strabox/caravela/util"
)

type Scheduler struct {
	config            *configuration.Configuration  // System's configuration
	discovery         apiInternal.DiscoveryInternal // Discovery module
	containersManager *containers.Manager           // Containers manager
	client            remote.Caravela               // Caravela's remote client
}

func NewScheduler(config *configuration.Configuration, internalDisc apiInternal.DiscoveryInternal,
	containersManager *containers.Manager, client remote.Caravela) *Scheduler {

	res := &Scheduler{}
	res.config = config
	res.discovery = internalDisc
	res.containersManager = containersManager
	res.client = client
	return res
}

/*
Executed when the local user wants to deploy a container in the system.
*/
func (scheduler *Scheduler) Deploy(containerImageKey string, containerArgs []string, cpus int, ram int) {
	log.Debugf(util.LogTag("[Deploy]")+"%s , CPUs: %d, RAM: %d", containerImageKey, cpus, ram)
	remoteNodes := scheduler.discovery.Find(*resources.NewResources(cpus, ram))
	log.Debugf(util.LogTag("[Deploy]")+"Remote Nodes: %v", remoteNodes)
	for _, v := range remoteNodes {
		log.Debugf(util.LogTag("[Deploy]")+"Node: %s", v.GUID().String())
		_, offers := scheduler.client.GetOffers(v.IPAddress(), v.GUID().String())
		log.Debugf(util.LogTag("[Deploy]")+"Offers: %v", offers)
		if len(offers) >= 1 {
			log.Debugf(util.LogTag("[Deploy]") + "Launching")
			firstOffer := offers[0]
			err := scheduler.client.LaunchContainer(firstOffer.SupplierIP, scheduler.config.HostIP(), firstOffer.ID,
				containerImageKey, containerArgs, cpus, ram)
			if err != nil {
				log.Debugf(util.LogTag("[Deploy]")+"Launch error: %v", err)
			}
			return
		}
	}
}

/*
Executed when a system's node wants to launch a container in this node.
*/
func (scheduler *Scheduler) Launch(fromBuyerIP string, offerID int64, containerImageKey string, containerArgs []string,
	cpus int, ram int) {
	log.Debugf(util.LogTag("[Launch]")+"%s , CPUs: %d, RAM: %d", containerImageKey, cpus, ram)
	scheduler.containersManager.StartContainer(fromBuyerIP, containerImageKey, containerArgs, offerID)
}
