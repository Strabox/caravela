package cli

import (
	"github.com/strabox/caravela/api"
	"github.com/strabox/caravela/api/remote"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/docker"
	"github.com/strabox/caravela/node"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/overlay/chord"
)

func initNode(hostIP string, join bool, joinIP string) error {
	var systemConfigurations *configuration.Configuration
	var err error = nil

	// Create configuration structures from the configuration file (if it exists)
	if join {
		caravelaClient := remote.NewHttpClient(configuration.Default(hostIP))

		systemConfigurations, err = caravelaClient.ObtainConfiguration(joinIP)
		if err != nil {
			return err
		}

		systemConfigurations, err = configuration.ObtainExternal(hostIP, systemConfigurations)
		if err != nil {
			return err
		}
	} else {
		systemConfigurations, err = configuration.ReadFromFile(hostIP)
		if err != nil {
			return err
		}
	}

	// Print/log the systemConfigurations values
	systemConfigurations.Print()

	// Global GUID size initialization
	guid.InitializeGUID(systemConfigurations.ChordHashSizeBits())

	// Create Overlay Component (Chord overlay initial)
	overlay := chord.NewChordOverlay(guid.SizeBytes(), systemConfigurations.HostIP(), systemConfigurations.OverlayPort(),
		systemConfigurations.ChordVirtualNodes(), systemConfigurations.ChordNumSuccessors(), systemConfigurations.ChordTimeout())

	// Create CARAVELA's Remote Client
	caravelaCli := remote.NewHttpClient(systemConfigurations)

	// Create Docker client
	dockerClient := docker.NewDockerClient(systemConfigurations)

	// Create the API server
	apiServer := api.NewServer(systemConfigurations.APIPort())

	// Create a CARAVELA Node passing all the external components and start it functions
	thisNode := node.NewNode(systemConfigurations, overlay, caravelaCli, dockerClient, apiServer)
	err = thisNode.Start(join, joinIP)
	if err != nil {
		return err
	}

	return nil
}
