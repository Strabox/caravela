package containers

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/docker"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/discovery/api"
	"github.com/strabox/caravela/util"
	"sync"
	"time"
)

/*
Containers manager responsible for interacting with the Docker daemon and managing all the interaction with the
deployed containers.
*/
type Manager struct {
	started      bool                         // Used to know if the manager already started to work
	config       *configuration.Configuration // CARAVELA node configuration
	dockerClient docker.Client                // Custom Docker's client
	supplier     api.DiscoveryInternal        // Node supplier API

	checkContainersTicker <-chan time.Time        // Ticker to check for containers status
	containersMutex       *sync.Mutex             // Mutex to control access to containers map
	containersMap         map[string][]*Container // Collection of deployed containers buyerIP->Container
}

func NewManager(config *configuration.Configuration,
	dockerClient docker.Client, supplier api.DiscoveryInternal) *Manager {
	manager := &Manager{
		started:               false,
		config:                config,
		dockerClient:          dockerClient,
		supplier:              supplier,
		checkContainersTicker: time.NewTicker(config.CheckContainersInterval()).C,
		containersMutex:       &sync.Mutex{},
		containersMap:         make(map[string][]*Container),
	}
	return manager
}

func (man *Manager) Start() {
	if !man.started {
		man.started = true
		go man.checkDeployedContainers()
	}
}

func (man *Manager) checkDeployedContainers() {
	for {
		select {
		case <-man.checkContainersTicker:
			man.containersMutex.Lock()
			log.Debug(util.LogTag("[Manager]") + "Checking containers...")

			for key, containers := range man.containersMap {

				for i := len(containers) - 1; i >= 0; i-- {
					contStatus, err := man.dockerClient.CheckContainerStatus(man.containersMap[key][i].DockerID())
					if err == nil && !contStatus.IsRunning() {
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
		}
	}
}

/*
Verify if the offer is valid and alert the supplier and after that start the container in the Docker engine.
*/
func (man *Manager) StartContainer(buyerIP string, imageKey string, args []string, offerID int64,
	resourcesNecessary resources.Resources) error {

	man.containersMutex.Lock()
	defer man.containersMutex.Unlock()

	obtained := man.supplier.ObtainResources(offerID, resourcesNecessary)
	if obtained {
		containerID, err := man.dockerClient.RunContainer(imageKey, args, "0", resourcesNecessary.RAM())
		if err == nil {
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
			man.supplier.ReturnResources(resourcesNecessary)
			return err
		}
	} else {
		return fmt.Errorf("offer couldn't be obtained in the local supplier")
	}
}
