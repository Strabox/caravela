package containers

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/docker/events"
	"github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/external"
	"github.com/strabox/caravela/util"
	"github.com/strabox/caravela/util/debug"
	"sync"
	"unsafe"
)

// Containers manager responsible for interacting with the Docker daemon and managing all the interaction with the
// deployed containers.
// Basically it is a local node manager for the containers.
type Manager struct {
	common.NodeComponent // Base component.

	config       *configuration.Configuration // System's configurations.
	dockerClient external.DockerClient        // Docker's client.
	supplier     supplierLocal                // Local Supplier component.

	quitChan        chan bool                             // Channel to alert that the node is stopping.
	containersMutex sync.Mutex                            // Mutex to control access to containers map.
	containersMap   map[string]map[string]*localContainer // Collection of deployed containers (buyerIP->(containerID->Container)).
}

// NewManager creates a new containers manager component.
func NewManager(config *configuration.Configuration, dockerClient external.DockerClient,
	supplier supplierLocal) *Manager {
	return &Manager{
		config:       config,
		dockerClient: dockerClient,
		supplier:     supplier,

		quitChan:        make(chan bool),
		containersMutex: sync.Mutex{},
		containersMap:   make(map[string]map[string]*localContainer),
	}
}

// receiveDockerEvents
func (m *Manager) receiveDockerEvents(eventsChan <-chan *events.Event) {
	go func() {
		for {
			select {
			case event := <-eventsChan:
				if event.Type == events.ContainerDied {
					m.StopContainer(event.Value)
				}
			case quit := <-m.quitChan: // Stopping the containers management
				if quit {
					log.Infof(util.LogTag("CONTAINER") + "STOPPED")
					return
				}
			}
		}
	}()
}

// Verify if the offer is valid and alert the supplier and after that start the container in the Docker engine.
func (m *Manager) StartContainer(fromBuyer *types.Node, offer *types.Offer, containersConfigs []types.ContainerConfig,
	totalResourcesNecessary resources.Resources) ([]types.ContainerStatus, error) {
	if !m.IsWorking() {
		panic(fmt.Errorf("can't start container, container manager not working"))
	}

	m.containersMutex.Lock()
	defer m.containersMutex.Unlock()

	// =================== Obtain the resources from the offer ==================

	obtained := m.supplier.ObtainResources(offer.ID, totalResourcesNecessary, len(containersConfigs))
	if !obtained {
		log.Debugf(util.LogTag("CONTAINER")+"Container NOT RUNNING, invalid offer: %d", offer.ID)
		return nil, fmt.Errorf("can't start container, invalid offer: %d", offer.ID)
	}

	// =================== Launch container in the Docker Engine ================

	deployedContStatus := make([]types.ContainerStatus, 0)

	for _, contConfig := range containersConfigs {
		containerStatus, err := m.dockerClient.RunContainer(contConfig)
		if err != nil { // If can't deploy a container remove all the other containers.
			m.supplier.ReturnResources(totalResourcesNecessary, len(containersConfigs))
			for _, contStatus := range deployedContStatus {
				m.StopContainer(contStatus.ContainerID)
			}
			return nil, err
		}
		deployedContStatus = append(deployedContStatus, *containerStatus)
	}

	// =================== Updates the inner container structures ================

	for i, contConfig := range containersConfigs {
		containerID := deployedContStatus[i].ContainerID
		contResources := resources.NewResourcesCPUClass(int(contConfig.Resources.CPUClass), contConfig.Resources.CPUs, contConfig.Resources.Memory)
		newContainer := newContainer(contConfig.Name, contConfig.ImageKey, contConfig.Args, contConfig.PortMappings,
			*contResources, containerID, fromBuyer.IP)

		if _, ok := m.containersMap[fromBuyer.IP]; !ok {
			userContainersMap := make(map[string]*localContainer)
			userContainersMap[containerID] = newContainer
			m.containersMap[fromBuyer.IP] = userContainersMap
		} else {
			m.containersMap[fromBuyer.IP][containerID] = newContainer
		}

		deployedContStatus[i].SupplierIP = m.config.HostIP() // Set the container's supplier's IP!

		log.Debugf(util.LogTag("CONTAINER")+"[%d] Container %s RUNNING, Img: %s, Args: %v, Res: <%d,%d>",
			i, containerID[0:12], contConfig.ImageKey, contConfig.Args, contResources.CPUs(),
			contResources.Memory())
	}

	return deployedContStatus, nil
}

// StopContainer stop a local container in the Docker engine and remove it.
func (m *Manager) StopContainer(containerIDToStop string) error {
	m.containersMutex.Lock()
	defer m.containersMutex.Unlock()

	for buyerIP, containersMap := range m.containersMap {
		for containerID, container := range containersMap {
			if containerID == containerIDToStop {
				m.dockerClient.RemoveContainer(containerIDToStop)
				m.supplier.ReturnResources(container.Resources(), 1)
				delete(containersMap, containerID)
				return nil
			}
		}
		if containersMap == nil || len(containersMap) == 0 {
			delete(m.containersMap, buyerIP)
		}
	}

	return errors.New("container does not exist")
}

// ===============================================================================
// =							SubComponent Interface                           =
// ===============================================================================

func (m *Manager) Start() {
	m.Started(m.config.Simulation(), func() {
		if !m.config.Simulation() {
			eventsChan := m.dockerClient.Start()
			m.receiveDockerEvents(eventsChan)
		}
	})
}

func (m *Manager) Stop() {
	m.Stopped(func() {
		m.containersMutex.Lock()
		defer m.containersMutex.Unlock()

		// Stop and remove all the running containers from the docker engine
		for _, containers := range m.containersMap {
			for containerID := range containers {
				m.dockerClient.RemoveContainer(containerID)
				log.Debugf(util.LogTag("CONTAINER")+"Container, %s STOPPED and REMOVED", containerID)
			}
		}

		m.quitChan <- true
	})
}

func (m *Manager) IsWorking() bool {
	return m.Working()
}

// ===============================================================================
// =							    Debug Methods                                =
// ===============================================================================

func (m *Manager) DebugSizeBytes() int {
	localContainerSize := func(container *localContainer) uintptr {
		contSizeBytes := unsafe.Sizeof(*container)
		contSizeBytes += debug.SizeofString(container.buyerIP)
		// common.Container
		contSizeBytes += unsafe.Sizeof(*container.Container)
		contSizeBytes += debug.SizeofString(container.Name())
		contSizeBytes += debug.SizeofString(container.ImageKey())
		contSizeBytes += debug.SizeofString(container.ID())
		contSizeBytes += debug.SizeofStringSlice(container.Args())
		contSizeBytes += debug.SizeofPortMappings(container.PortMappings())
		return contSizeBytes
	}

	contManagerSizeBytes := unsafe.Sizeof(*m)
	for k, v := range m.containersMap {
		contManagerSizeBytes += unsafe.Sizeof(k)
		contManagerSizeBytes += debug.SizeofString(k)
		contManagerSizeBytes += unsafe.Sizeof(v)
		if v != nil {
			for k2, v2 := range v {
				contManagerSizeBytes += unsafe.Sizeof(k2)
				contManagerSizeBytes += debug.SizeofString(k2)
				contManagerSizeBytes += unsafe.Sizeof(v2)
				contManagerSizeBytes += localContainerSize(v2)
			}
		}
	}
	return int(contManagerSizeBytes)
}
