package scheduler

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/api/remote"
	"github.com/strabox/caravela/api/rest"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/containers"
	apiInternal "github.com/strabox/caravela/node/discovery/api"
	"github.com/strabox/caravela/util"
)

/*
Scheduler entity responsible for receiving local and remote requests for deploying containers
running in the system. It takes a request for running a container and decides where to deploy it
in conjunction with the Discovery module.
*/
type Scheduler struct {
	common.SystemSubComponent // Base component

	config *configuration.Configuration // System's configuration
	client remote.Caravela              // Caravela's remote client

	discovery         apiInternal.DiscoveryInternal // Discovery module
	containersManager *containers.Manager           // Containers manager module
}

func NewScheduler(config *configuration.Configuration, internalDisc apiInternal.DiscoveryInternal,
	containersManager *containers.Manager, client remote.Caravela) *Scheduler {

	res := &Scheduler{}
	res.config = config
	res.client = client

	res.discovery = internalDisc
	res.containersManager = containersManager
	return res
}

/*
Executed when the local user wants to deploy a container in the system.
*/
func (scheduler *Scheduler) Run(containerImageKey string, portMappings []rest.PortMapping, containerArgs []string,
	cpus int, ram int) error {

	if !scheduler.isWorking() {
		panic(fmt.Errorf("can't run container, scheduler not working"))
	}
	log.Debugf(util.LogTag("Deploy")+"Deploying... %s , CPUs: %d, RAM: %d", containerImageKey, cpus, ram)

	offers := scheduler.discovery.FindOffers(*resources.NewResources(cpus, ram))

	for _, offer := range offers {
		err := scheduler.client.LaunchContainer(offer.SupplierIP, scheduler.config.HostIP(), offer.ID,
			containerImageKey, portMappings, containerArgs, cpus, ram)
		if err == nil {
			log.Debugf(util.LogTag("Deploy")+"Deployed %s , CPUs: %d, RAM: %d", containerImageKey, cpus, ram)
			return nil
		} else {
			log.Debugf(util.LogTag("Deploy")+"Deploy error: %v", err)
		}
	}
	// TODO: Try more offers because they can exist in the system.
	return fmt.Errorf("no offers found to deploy the container")
}

/*
Executed when a system's node wants to launch a container in this node.
*/
func (scheduler *Scheduler) Launch(fromBuyerIP string, offerID int64, containerImageKey string,
	portMappings []rest.PortMapping, containerArgs []string, cpus int, ram int) error {

	if !scheduler.isWorking() {
		panic(fmt.Errorf("can't launch container, scheduler not working"))
	}
	log.Debugf(util.LogTag("Launch")+"Launching %s , CPUs: %d, RAM: %d ...", containerImageKey, cpus, ram)

	resourcesNecessary := resources.NewResources(cpus, ram)
	err := scheduler.containersManager.StartContainer(fromBuyerIP, containerImageKey, portMappings,
		containerArgs, offerID, *resourcesNecessary)

	return err
}

func (scheduler *Scheduler) Start() {
	scheduler.Started(func() {})
}

func (scheduler *Scheduler) Stop() {
	scheduler.Stopped(func() {})
}

func (scheduler *Scheduler) isWorking() bool {
	return scheduler.Working()
}
