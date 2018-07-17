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

// Scheduler is responsible for receiving local and remote requests for deploying containers
// running in the system. It takes a request for running a container and decides where to deploy it
// in conjunction with the Discovery module.
type Scheduler struct {
	common.NodeComponent // Base component

	config *configuration.Configuration // System's configuration
	client remote.Caravela              // Caravela's remote client

	discovery         apiInternal.DiscoveryInternal // Local Discovery module
	containersManager *containers.Manager           // Containers manager module
}

func NewScheduler(config *configuration.Configuration, internalDisc apiInternal.DiscoveryInternal,
	containersManager *containers.Manager, client remote.Caravela) *Scheduler {

	return &Scheduler{
		config:            config,
		client:            client,
		discovery:         internalDisc,
		containersManager: containersManager,
	}
}

// Executed when a system's node wants to launch a container in this node.
func (scheduler *Scheduler) Launch(fromBuyerIP string, offerID int64, containerImageKey string,
	portMappings []rest.PortMapping, containerArgs []string, cpus int, ram int) (string, error) {

	if !scheduler.isWorking() {
		panic(fmt.Errorf("can't launch container, scheduler not working"))
	}
	log.Debugf(util.LogTag("Launch")+"Launching %s , Resources: <%d,%d> ...", containerImageKey, cpus, ram)

	resourcesNecessary := resources.NewResources(cpus, ram)
	containerID, err := scheduler.containersManager.StartContainer(fromBuyerIP, containerImageKey, portMappings,
		containerArgs, offerID, *resourcesNecessary)

	return containerID, err
}

func (scheduler *Scheduler) SubmitContainers(containerImageKey string, portMappings []rest.PortMapping, containerArgs []string,
	cpus int, ram int) (string, string, error) {

	if !scheduler.isWorking() {
		panic(fmt.Errorf("can't run container, scheduler not working"))
	}
	log.Debugf(util.LogTag("Run")+"Deploying... %s , CPUs: %d, RAM: %d", containerImageKey, cpus, ram)

	offers := scheduler.discovery.FindOffers(*resources.NewResources(cpus, ram))

	for _, offer := range offers {
		contStatus, err := scheduler.client.LaunchContainer(offer.SupplierIP, scheduler.config.HostIP(), offer.ID,
			containerImageKey, portMappings, containerArgs, cpus, ram)
		if err != nil {
			log.Debugf(util.LogTag("Run")+"Deploy error: %s", err)
			continue
		}

		log.Debugf(util.LogTag("Run")+"Deployed %s , CPUs: %d, RAM: %d", containerImageKey, cpus, ram)
		return contStatus.ID, offer.SupplierIP, nil
	}

	log.Debugf(util.LogTag("Run") + "No offers found")
	return "", "", fmt.Errorf("no offers found to deploy the container")
}

/*
===============================================================================
							SubComponent Interface
===============================================================================
*/

func (scheduler *Scheduler) Start() {
	scheduler.Started(scheduler.config.Simulation(), func() { /* Do Nothing */ })
}

func (scheduler *Scheduler) Stop() {
	scheduler.Stopped(func() { /* Do Nothing */ })
}

func (scheduler *Scheduler) isWorking() bool {
	return scheduler.Working()
}
