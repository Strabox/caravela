package user

import (
	"context"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/util"
	"sync"
)

type Manager struct {
	common.NodeComponent // Base component

	containers          sync.Map // Map ID<->Container submitted by the user
	minRequestResources resources.Resources
	localScheduler      localScheduler   // Container's scheduler component
	userRemoteCli       userRemoteClient //

	config *configuration.Configuration // System's configurations.
}

func NewManager(config *configuration.Configuration, localScheduler localScheduler, userRemoteCli userRemoteClient,
	minRequestResources resources.Resources) *Manager {
	return &Manager{
		minRequestResources: minRequestResources,
		config:              config,
		localScheduler:      localScheduler,
		userRemoteCli:       userRemoteCli,

		containers: sync.Map{},
	}
}

func (man *Manager) SubmitContainers(ctx context.Context, containerConfigs []types.ContainerConfig) error {
	// Validate request
	for i, contConfig := range containerConfigs {
		if contConfig.Resources.RAM == 0 && contConfig.Resources.CPUs == 0 {
			containerConfigs[i].Resources.CPUs = man.minRequestResources.CPUs()
			containerConfigs[i].Resources.RAM = man.minRequestResources.RAM()
		} else if contConfig.Resources.RAM == 0 || contConfig.Resources.CPUs == 0 {
			return fmt.Errorf("invalid resources resquest")
		}
	}

	// Contact local scheduler to submit the request into the system
	containersStatus, err := man.localScheduler.SubmitContainers(ctx, containerConfigs)
	if err != nil {
		return err
	}

	// Update internals
	for _, contStatus := range containersStatus {
		container := newContainer(contStatus.Name, contStatus.ImageKey, contStatus.Args, contStatus.PortMappings,
			*resources.NewResources(contStatus.Resources.CPUs, contStatus.Resources.RAM), contStatus.ContainerID,
			contStatus.SupplierIP)

		man.containers.Store(container.ShortID(), container)
	}

	return nil
}

func (man *Manager) StopContainers(ctx context.Context, containerIDs []string) error {
	errMsg := "Failed to stop:"
	fail := false
	for _, contID := range containerIDs {
		contTmp, contExist := man.containers.Load(contID)
		container, ok := contTmp.(*deployedContainer)
		if contExist && ok {
			if err := man.userRemoteCli.StopLocalContainer(ctx, &types.Node{IP: container.supplierIP()}, container.ID()); err == nil {
				man.containers.Delete(contID)
			} else {
				fail = true
				errMsg += " " + contID
			}
		}
	}

	if fail {
		err := errors.New(errMsg)
		log.Debugf(util.LogTag("USRMNG")+"STOPPING containers, error: %s", err)
		return err
	}

	return nil
}

func (man *Manager) ListContainers() []types.ContainerStatus {
	res := make([]types.ContainerStatus, 0)

	man.containers.Range(func(key, value interface{}) bool {
		if container, ok := value.(*deployedContainer); ok {
			res = append(res,
				types.ContainerStatus{
					ContainerConfig: types.ContainerConfig{
						Name:         container.Name(),
						ImageKey:     container.ImageKey(),
						PortMappings: container.PortMappings(),
					},
					SupplierIP:  container.supplierIP(),
					ContainerID: container.ShortID(),
					Status:      "Running", // TODO: Solve this hardcode
				})
		}
		return true
	})
	return res
}

// ===============================================================================
// =							SubComponent Interface                           =
// ===============================================================================

func (man *Manager) Start() {
	man.Started(man.config.Simulation(), func() { /* Do Nothing */ })
}

func (man *Manager) Stop() {
	man.Stopped(func() { /* Do Nothing */ })
}

func (man *Manager) isWorking() bool {
	return man.Working()
}
