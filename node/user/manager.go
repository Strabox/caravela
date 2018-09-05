package user

import (
	"context"
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

func (m *Manager) SubmitContainers(ctx context.Context, containerConfigs []types.ContainerConfig) ([]types.ContainerStatus, error) {
	// Validate container submission request.
	coLocationPolicy := -1
	for i, contConfig := range containerConfigs {
		// If a resource constraint is specified to 0 (user does not care) we use the minimum resources in our partitions.
		if contConfig.Resources.CPUClass == 0 {
			containerConfigs[i].Resources.CPUClass = types.CPUClass(m.minRequestResources.CPUClass())
		} else if contConfig.Resources.CPUs == 0 {
			containerConfigs[i].Resources.CPUs = m.minRequestResources.CPUs()
		} else if contConfig.Resources.Memory == 0 {
			containerConfigs[i].Resources.Memory = m.minRequestResources.Memory()
		}

		// Containers with co-location group policies must have the same CPU Class specified.
		if contConfig.GroupPolicy == types.CoLocationGroupPolicy && coLocationPolicy == -1 {
			coLocationPolicy = int(contConfig.GroupPolicy)
		} else if contConfig.GroupPolicy == types.CoLocationGroupPolicy &&
			coLocationPolicy != -1 && int(contConfig.GroupPolicy) != coLocationPolicy {
			return nil, errors.New("containers with co-location policies must have the same CPU Class constraint")
		}
	}

	// Submit the request into the local scheduler.
	containersStatus, err := m.localScheduler.SubmitContainers(ctx, containerConfigs)
	if err != nil {
		return nil, err
	}

	// Update internals.
	for _, contStatus := range containersStatus {
		container := newContainer(contStatus.Name, contStatus.ImageKey, contStatus.Args, contStatus.PortMappings,
			*resources.NewResourcesCPUClass(int(contStatus.Resources.CPUClass), contStatus.Resources.CPUs, contStatus.Resources.Memory), contStatus.ContainerID,
			contStatus.SupplierIP)

		m.containers.Store(container.ShortID(), container)
	}

	return containersStatus, nil
}

func (m *Manager) StopContainers(ctx context.Context, containerIDs []string) error {
	errMsg := "Failed to stop:"
	fail := false
	for _, contID := range containerIDs {
		contTmp, contExist := m.containers.Load(contID[:common.ContainerShortIDSize])
		container, ok := contTmp.(*deployedContainer)
		if contExist && ok {
			if err := m.userRemoteCli.StopLocalContainer(ctx, &types.Node{IP: container.supplierIP()}, container.ID()); err == nil {
				m.containers.Delete(contID)
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

func (m *Manager) ListContainers() []types.ContainerStatus {
	res := make([]types.ContainerStatus, 0)

	m.containers.Range(func(_, value interface{}) bool {
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

func (m *Manager) Start() {
	m.Started(m.config.Simulation(), func() { /* Do Nothing */ })
}

func (m *Manager) Stop() {
	m.Stopped(func() { /* Do Nothing */ })
}

func (m *Manager) isWorking() bool {
	return m.Working()
}
