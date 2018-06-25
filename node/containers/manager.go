package containers

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/docker"
	"github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/discovery/api"
	"github.com/strabox/caravela/util"
	"sync"
	"time"
)

/*
Containers manager responsible for interacting with the Docker daemon and managing all the interaction with the
deployed containers.
Basically it is a local node manager for the containers.
*/
type Manager struct {
	common.SystemSubComponent // Base component

	config       *configuration.Configuration // System's configuration
	dockerClient docker.Client                // Docker's client
	supplier     api.DiscoveryInternal        // Node supplier API

	quitChan              chan bool               // Channel to alert that the node is stopping
	checkContainersTicker <-chan time.Time        // Ticker to check for containers status
	containersMutex       *sync.Mutex             // Mutex to control access to containers map
	containersMap         map[string][]*Container // Collection of deployed containers buyerIP->Container
}

func NewManager(config *configuration.Configuration,
	dockerClient docker.Client, supplier api.DiscoveryInternal) *Manager {
	manager := &Manager{
		config:                config,
		dockerClient:          dockerClient,
		supplier:              supplier,
		quitChan:              make(chan bool),
		checkContainersTicker: time.NewTicker(config.CheckContainersInterval()).C,
		containersMutex:       &sync.Mutex{},
		containersMap:         make(map[string][]*Container),
	}
	return manager
}

func (man *Manager) checkDeployedContainers() {
	for {
		select {
		case <-man.checkContainersTicker: // Checking the submitted containers status
			man.containersMutex.Lock()

			for key, containers := range man.containersMap {

				for i := len(containers) - 1; i >= 0; i-- {
					containerID := man.containersMap[key][i].DockerID()
					contStatus, err := man.dockerClient.CheckContainerStatus(containerID)
					if err == nil && !contStatus.IsRunning() {
						log.Debugf(util.LogTag("ContManager")+"Container, %s STOPPED", containerID)
						man.dockerClient.RemoveContainer(containers[i].dockerID)
						man.supplier.ReturnResources(containers[i].resources)
						man.containersMap[key] = append(man.containersMap[key][:i], man.containersMap[key][i+1:]...)
					}
				}

				if man.containersMap[key] == nil || len(man.containersMap[key]) == 0 {
					delete(man.containersMap, key)
				}

			}
			man.containersMutex.Unlock()
		case res := <-man.quitChan: // Stopping the containers management
			if res {
				log.Infof(util.LogTag("ContManager") + "STOPPED")
				return
			}
		}
	}
}

/*
Verify if the offer is valid and alert the supplier and after that start the container in the Docker engine.
*/
func (man *Manager) StartContainer(buyerIP string, imageKey string, args []string, offerID int64,
	resourcesNecessary resources.Resources) error {

	if !man.isWorking() {
		panic(fmt.Errorf("can't start container, container manager not working"))
	}
	man.containersMutex.Lock()
	defer man.containersMutex.Unlock()

	if !man.isWorking() {
		panic(fmt.Errorf("can't start container, containers manager not working"))
	}

	obtained := man.supplier.ObtainResources(offerID, resourcesNecessary)
	if obtained {
		containerID, err := man.dockerClient.RunContainer(imageKey, args, "0", resourcesNecessary.RAM())
		if err == nil {
			log.Debugf(util.LogTag("ContManager")+"Container %s RUNNING, Image: %s , Args: %v, Resources: <%d,%d>",
				containerID, imageKey, args, resourcesNecessary.CPUs(), resourcesNecessary.RAM())
			newContainer := NewContainer(containerID, buyerIP, resourcesNecessary)
			if man.containersMap[buyerIP] == nil {
				containersList := make([]*Container, 1)
				containersList[0] = newContainer
				man.containersMap[buyerIP] = containersList
			} else {
				man.containersMap[buyerIP] = append(man.containersMap[buyerIP], newContainer)
			}
			return nil
		} else {
			log.Errorf(util.LogTag("ContManager")+"Can't start container, docker error: %s", err)
			man.supplier.ReturnResources(resourcesNecessary)
			return err
		}
	} else {
		log.Debugf(util.LogTag("ContManager")+"Can't start container, invalid offer: %d", offerID)
		return fmt.Errorf("can't start container: invalid offer: %d", offerID)
	}
}

func (man *Manager) Start() {
	man.Started(func() {
		go man.checkDeployedContainers()
	})
}

func (man *Manager) Stop() {
	man.Stopped(func() {
		// TODO clean up all running containers
		man.quitChan <- true
	})
}

func (man *Manager) isWorking() bool {
	return man.Working()
}
