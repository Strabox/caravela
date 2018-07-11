package user

import (
	"github.com/strabox/caravela/api/rest"
	"github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/resources"
	"sync"
)

type Manager struct {
	common.SystemSubComponent // Base component

	localScheduler localScheduler // Scheduler component

	containers sync.Map // Map ID<->Container submitted by the user
}

func NewManager(localScheduler localScheduler) *Manager {
	return &Manager{
		localScheduler: localScheduler,

		containers: sync.Map{},
	}
}

func (man *Manager) SubmitContainers(containerImageKey string, portMappings []rest.PortMapping,
	containerArgs []string, cpus int, ram int) error {

	containerID, suppIP, err := man.localScheduler.SubmitContainers(containerImageKey, portMappings, containerArgs, cpus, ram)
	if err != nil {
		return err
	}

	man.containers.Store(containerID, newContainer(containerImageKey, containerArgs, portMappings,
		*resources.NewResources(cpus, ram), containerID, suppIP))

	return nil
}

func (man *Manager) StopContainers(containerIDs []string) error {
	err := man.localScheduler.StopContainers(containerIDs)
	if err != nil {
		return err
	}

	for _, contID := range containerIDs {
		man.containers.Delete(contID)
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
