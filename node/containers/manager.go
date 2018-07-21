package containers

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/discovery/api"
	"github.com/strabox/caravela/node/external"
	"github.com/strabox/caravela/util"
	"sync"
	"time"
)

// Containers manager responsible for interacting with the Docker daemon and managing all the interaction with the
// deployed containers.
// Basically it is a local node manager for the containers.
type Manager struct {
	common.NodeComponent // Base component

	config       *configuration.Configuration // System's configuration
	dockerClient external.DockerClient        // Docker's client
	supplier     api.DiscoveryInternal        // Node supplier API

	quitChan              chan bool                             // Channel to alert that the node is stopping
	checkContainersTicker <-chan time.Time                      // Ticker to check for containers status
	containersMutex       *sync.Mutex                           // Mutex to control access to containers map
	containersMap         map[string]map[string]*localContainer // Collection of deployed containers (buyerIP->(containerID->Container))
}

// NewManager creates a new containers manager component.
func NewManager(config *configuration.Configuration, dockerClient external.DockerClient,
	supplier api.DiscoveryInternal) *Manager {
	return &Manager{
		config:       config,
		dockerClient: dockerClient,
		supplier:     supplier,

		quitChan:              make(chan bool),
		checkContainersTicker: time.NewTicker(config.CheckContainersInterval()).C,
		containersMutex:       &sync.Mutex{},
		containersMap:         make(map[string]map[string]*localContainer),
	}
}

func (man *Manager) checkDeployedContainers() {
	for {
		select {
		case <-man.checkContainersTicker: // Checking the submitted containers status and remove them if they finished
			go func() {
				man.containersMutex.Lock()
				defer man.containersMutex.Unlock()

				for key, containerMap := range man.containersMap {

					for containerID, container := range containerMap {
						contStatus, err := man.dockerClient.CheckContainerStatus(containerID)
						if err == nil && !contStatus.IsRunning() {
							log.Debugf(util.LogTag("CONTAINER")+"Container, %s STOPPED and REMOVED", containerID)
							man.dockerClient.RemoveContainer(containerID)
							man.supplier.ReturnResources(container.Resources())
							delete(containerMap, containerID)
						}
					}

					if containerMap == nil || len(containerMap) == 0 {
						delete(man.containersMap, key)
					}
				}
			}()
		case res := <-man.quitChan: // Stopping the containers management
			if res {
				log.Infof(util.LogTag("CONTAINER") + "STOPPED")
				return
			}
		}
	}
}

// Verify if the offer is valid and alert the supplier and after that start the container in the Docker engine.
func (man *Manager) StartContainer(fromBuyer *types.Node, offer *types.Offer, containerConfig *types.ContainerConfig,
	resourcesNecessary resources.Resources) (*types.ContainerStatus, error) {
	if !man.isWorking() {
		panic(fmt.Errorf("can't start container, container manager not working"))
	}

	man.containersMutex.Lock()
	defer man.containersMutex.Unlock()

	obtained := man.supplier.ObtainResources(offer.ID, resourcesNecessary)
	if !obtained {
		log.Debugf(util.LogTag("CONTAINER")+"Container NOT RUNNING, invalid offer: %v", offer)
		return nil, fmt.Errorf("can't start container, invalid offer: %v", offer)
	}

	containerID, err := man.dockerClient.RunContainer(containerConfig.ImageKey, containerConfig.PortMappings,
		containerConfig.Args, int64(resourcesNecessary.CPUs()), resourcesNecessary.RAM())
	if err != nil {
		man.supplier.ReturnResources(resourcesNecessary)
		return nil, err
	}

	newContainer := newContainer(containerConfig.ImageKey, containerConfig.Args, containerConfig.PortMappings,
		resourcesNecessary, containerID, fromBuyer.IP)
	if man.containersMap[fromBuyer.IP] == nil {
		userContainersMap := make(map[string]*localContainer)
		userContainersMap[containerID] = newContainer
		man.containersMap[fromBuyer.IP] = userContainersMap
	} else {
		man.containersMap[fromBuyer.IP][containerID] = newContainer
	}

	log.Debugf(util.LogTag("CONTAINER")+"Container %s RUNNING, Img: %s, Args: %v, Res: <%d,%d>",
		containerID[0:12], containerConfig.ImageKey, containerConfig.Args, resourcesNecessary.CPUs(),
		resourcesNecessary.RAM())

	return &types.ContainerStatus{
		ContainerConfig: *containerConfig,
		ContainerID:     containerID,
		Status:          "Running",
	}, nil
}

// Stop a local container and remove it.
func (man *Manager) StopContainer(containerIDToStop string) error {
	man.containersMutex.Lock()
	defer man.containersMutex.Unlock()

	for _, containersMap := range man.containersMap {
		for containerID, container := range containersMap {
			if containerID == containerIDToStop {
				man.dockerClient.RemoveContainer(containerIDToStop)
				man.supplier.ReturnResources(container.Resources())
				delete(containersMap, containerID)
				return nil
			}
		}
	}

	return fmt.Errorf("container does not exist")
}

/*
===============================================================================
							SubComponent Interface
===============================================================================
*/

func (man *Manager) Start() {
	man.Started(man.config.Simulation(), func() {
		if !man.config.Simulation() {
			go man.checkDeployedContainers()
		}
	})
}

func (man *Manager) Stop() {
	man.Stopped(func() {
		man.containersMutex.Lock()
		defer man.containersMutex.Unlock()

		// Stop and remove all the running containers from the docker engine
		for _, containers := range man.containersMap {
			for containerID := range containers {
				man.dockerClient.RemoveContainer(containerID)
				log.Debugf(util.LogTag("CONTAINER")+"Container, %s STOPPED and REMOVED", containerID)
			}
		}

		man.quitChan <- true
	})
}

func (man *Manager) isWorking() bool {
	return man.Working()
}
