package scheduler

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/containers"
	apiInternal "github.com/strabox/caravela/node/discovery/api"
	"github.com/strabox/caravela/node/external"
	"github.com/strabox/caravela/util"
)

// Scheduler is responsible for receiving local and remote requests for deploying containers
// running in the system. It takes a request for running a container and decides where to deploy it
// in conjunction with the Discovery module.
type Scheduler struct {
	common.NodeComponent // Base component

	config *configuration.Configuration // System's configuration
	client external.Caravela            // Caravela's remote client

	discovery         apiInternal.DiscoveryInternal // Local Discovery module
	containersManager *containers.Manager           // Containers manager module
}

func NewScheduler(config *configuration.Configuration, internalDisc apiInternal.DiscoveryInternal,
	containersManager *containers.Manager, client external.Caravela) *Scheduler {

	return &Scheduler{
		config:            config,
		client:            client,
		discovery:         internalDisc,
		containersManager: containersManager,
	}
}

// Executed when a system's node wants to launch a container in this node.
func (scheduler *Scheduler) Launch(fromBuyer *types.Node, offer *types.Offer,
	containerConfig *types.ContainerConfig) (*types.ContainerStatus, error) {

	if !scheduler.isWorking() {
		panic(fmt.Errorf("can't launch container, scheduler not working"))
	}
	log.Debugf(util.LogTag("SCHEDULE")+"Launching... Img: %s, Res: <%d,%d>", containerConfig.ImageKey,
		containerConfig.Resources.CPUs, containerConfig.Resources.RAM)

	resourcesNecessary := resources.NewResources(containerConfig.Resources.CPUs, containerConfig.Resources.RAM)
	containerStatus, err := scheduler.containersManager.StartContainer(fromBuyer, offer, containerConfig, *resourcesNecessary)
	return containerStatus, err
}

func (scheduler *Scheduler) SubmitContainers(containerImageKey string, portMappings []types.PortMapping, containerArgs []string,
	cpus int, ram int) (string, string, error) {

	if !scheduler.isWorking() {
		panic(fmt.Errorf("can't run container, scheduler not working"))
	}
	log.Debugf(util.LogTag("SCHEDULE")+"Deploying... Img: %s , Res: <%d;%d>", containerImageKey, cpus, ram)

	offers := scheduler.discovery.FindOffers(*resources.NewResources(cpus, ram))

	for _, offer := range offers {
		log.Debugf(util.LogTag("SCHEDULE")+"Trying OFFER... SuppIP: %s, Offer: %d, Amount %d, Res: <%d;%d>",
			offer.SupplierIP, offer.ID, offer.Amount, offer.Resources.CPUs, offer.Resources.RAM)

		contStatus, err := scheduler.client.LaunchContainer(
			&types.Node{IP: scheduler.config.HostIP()},
			&types.Node{IP: offer.SupplierIP},
			&types.Offer{ID: offer.ID},
			&types.ContainerConfig{
				ImageKey:     containerImageKey,
				PortMappings: portMappings,
				Args:         containerArgs,
				Resources:    types.Resources{CPUs: cpus, RAM: ram},
			},
		)
		if err != nil {
			log.Debugf(util.LogTag("SCHEDULE")+"Deploy FAILED Offer: %d error: %s", offer.ID, err)
			continue
		}

		log.Debugf(util.LogTag("SCHEDULE")+"Deploy SUCCESS Img: %s, Res: <%d,%d>", containerImageKey, cpus, ram)
		return contStatus.ContainerID, offer.SupplierIP, nil
	}

	log.Debugf(util.LogTag("SCHEDULE") + "Deploy FAILED. No offers found.")
	return "", "", fmt.Errorf("no offers found to deploy")
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
