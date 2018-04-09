package containers

import (
	"github.com/strabox/caravela/docker"
	"github.com/strabox/caravela/node/discovery/api"
)

type Manager struct {
	dockerClient *docker.Client // Docker's client to interact with docker engine
	supplier     api.DiscoveryInternal
}

func NewManager(dockerClient *docker.Client, supplier api.DiscoveryInternal) *Manager {
	manager := &Manager{}
	manager.dockerClient = dockerClient
	manager.supplier = supplier
	return manager
}

func (m *Manager) RunContainer(imageKey string, args []string, offerID int64) error {
	err, offerRes := m.supplier.ObtainResourcesSlot(offerID)
	if err == nil {
		m.dockerClient.RunContainer(imageKey, args, "0", offerRes.RAM())
		return nil
	} else {
		return err
	}
}
