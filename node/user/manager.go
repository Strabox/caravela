package user

import (
	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/strabox/caravela/api/rest"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/util"
	"sync"
)

type Manager struct {
	common.NodeComponent // Base component

	config         *configuration.Configuration
	localScheduler localScheduler // Scheduler component
	userRemoteCli  userRemoteClient

	containers sync.Map // Map ID<->Container submitted by the user
}

func NewManager(config *configuration.Configuration, localScheduler localScheduler, userRemoteCli userRemoteClient) *Manager {
	return &Manager{
		config:         config,
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

	container := newContainer(containerImageKey, containerArgs, portMappings, *resources.NewResources(cpus, ram),
		containerID, suppIP)
	man.containers.Store(container.ShortID(), container)
	return nil
}

func (man *Manager) StopContainers(containerIDs []string) error {
	errMsg := "Failed to stop:"
	fail := false
	for _, contID := range containerIDs {
		contTmp, contExist := man.containers.Load(contID)
		container, ok := contTmp.(*deployedContainer)
		if contExist && ok {
			if err := man.userRemoteCli.StopLocalContainer(container.supplierIP(), container.ID()); err == nil {
				man.containers.Delete(contID)
			} else {
				fail = true
				errMsg += " " + contID
			}
		}
	}

	if fail {
		err := errors.New(errMsg)
		log.Debugf(util.LogTag("UsrMng")+" Error stopping containers: %s", err)
		return err
	}

	return nil
}

func (man *Manager) ListContainers() rest.ContainersList {
	res := rest.ContainersList{
		ContainersStatus: make([]rest.ContainerStatus, 0),
	}
	man.containers.Range(func(key, value interface{}) bool {
		if cont, ok := value.(*deployedContainer); ok {
			res.ContainersStatus = append(res.ContainersStatus,
				rest.ContainerStatus{
					ImageKey:     cont.ImageKey(),
					ID:           cont.ShortID(),
					PortMappings: cont.PortMappings(),
					Status:       "Running", // TODO: Solve this hardcode
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
	man.Started(man.config.Simulation(), func() { /* Do Nothing */ })
}

func (man *Manager) Stop() {
	man.Stopped(func() { /* Do Nothing */ })
}

func (man *Manager) isWorking() bool {
	return man.Working()
}
