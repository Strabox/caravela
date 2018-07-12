package user

import (
	"fmt"
	"github.com/strabox/caravela/api/rest"
	"github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/resources"
	"sync"
)

type Manager struct {
	common.SystemSubComponent // Base component

	localScheduler localScheduler // Scheduler component
	userRemoteCli  userRemoteClient

	containers sync.Map // Map ID<->Container submitted by the user
}

func NewManager(localScheduler localScheduler, userRemoteCli userRemoteClient) *Manager {
	return &Manager{
		localScheduler: localScheduler,
		userRemoteCli:  userRemoteCli,

		containers: sync.Map{},
	}
}

func (man *Manager) SubmitContainers(containerImageKey string, portMappings []rest.PortMapping,
	containerArgs []string, cpus int, ram int) error {

	containerID, suppIP, err := man.localScheduler.SubmitContainers(containerImageKey, portMappings,
		containerArgs, cpus, ram)
	if err != nil {
		return err
	}

	man.containers.Store(containerID, newContainer(containerImageKey, containerArgs, portMappings,
		*resources.NewResources(cpus, ram), containerID, suppIP))

	return nil
}

func (man *Manager) StopContainers(containerIDs []string) error {
	var errMsg = "Failed to stop:"
	var fail = false
	for _, contID := range containerIDs {
		contTmp, contExist := man.containers.Load(contID)
		container, ok := contTmp.(*container)
		if contExist && ok {
			if err := man.userRemoteCli.StopLocalContainer(container.supplierIP(), container.ID()); err == nil {
				man.containers.Delete(contID)
			} else {
				errMsg += " " + contID
			}
		}
	}

	if fail {
		return fmt.Errorf(errMsg)
	}

	return nil
}

func (man *Manager) ListContainers() rest.ContainersList {
	res := rest.ContainersList{
		ContainersStatus: make([]rest.ContainerStatus, 0),
	}
	man.containers.Range(func(key, value interface{}) bool {
		if cont, ok := value.(*container); ok {
			res.ContainersStatus = append(res.ContainersStatus,
				rest.ContainerStatus{
					ID: cont.ID(),
				})
		}
		return true
	})
	return res
}

/*
===============================================================================
							SubComponent Interface
===============================================================================
*/

func (man *Manager) Start() {
	man.Started(func() { /* Do Nothing */ })
}

func (man *Manager) Stop() {
	man.Stopped(func() { /* Do Nothing */ })
}

func (man *Manager) isWorking() bool {
	return man.Working()
}
