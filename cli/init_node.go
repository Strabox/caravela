package cli

import (
	"context"
	"fmt"
	"github.com/strabox/caravela/api"
	"github.com/strabox/caravela/api/remote"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/docker"
	"github.com/strabox/caravela/node"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/overlay/chord"
	"strings"
)

func initNode(hostIP, configFilePath string, join bool, joinIP string) error {
	var systemConfigurations *configuration.Configuration
	var err error = nil

	// Create configuration structures from the configuration file (if it exists)
	if join {
		caravelaClient := remote.NewClient(configuration.Default(hostIP))

		systemConfigurations, err = caravelaClient.ObtainConfiguration(context.Background(), &types.Node{IP: joinIP})
		if err != nil {
			return err
		}

		systemConfigurations, err = configuration.ObtainExternal(hostIP, systemConfigurations)
		if err != nil {
			return err
		}
	} else {
		systemConfigurations, err = configuration.ReadFromFile(hostIP, configFilePath)
		if err != nil && systemConfigurations != nil && strings.Contains(err.Error(), "cannot find the file") {
			fmt.Println("Information: using the default configuration")
		} else if err != nil && systemConfigurations == nil {
			return err
		}
	}

	// Print/log the systemConfigurations values
	systemConfigurations.Print()

	// Global GUID size initialization
	guid.Init(systemConfigurations.ChordHashSizeBits())

	// Create Overlay Component (Chord overlay initial)
	overlay := chord.New(guid.SizeBytes(), systemConfigurations.HostIP(), systemConfigurations.OverlayPort(),
		systemConfigurations.ChordVirtualNodes(), systemConfigurations.ChordNumSuccessors(), systemConfigurations.ChordTimeout())

	// Create CARAVELA's Remote Client
	caravelaCli := remote.NewClient(systemConfigurations)

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
