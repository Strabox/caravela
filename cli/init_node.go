package cli

import (
	"context"
	"fmt"
	"github.com/strabox/caravela/api/remote"
	"github.com/strabox/caravela/api/rest"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/docker"
	"github.com/strabox/caravela/node"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/overlay"
	"strings"
)

func initNode(hostIP, configFilePath string, join bool, joinIP string) error {
	var systemConfigurations *configuration.Configuration
	var err error = nil

	// Create configuration structures from the configuration file (if it exists)
	if join {
		defaultConfigs := configuration.Default(hostIP)
		caravelaClient := remote.NewClient(defaultConfigs.APIPort(), defaultConfigs.APITimeout())

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

	// GUID package custom initialization.
	guid.Init(systemConfigurations.ChordHashSizeBits(), int64(systemConfigurations.GUIDEstimatedNetworkSize()),
		int64(systemConfigurations.GUIDScaleFactor()))

	// Create Overlay Component
	overlayConfigured := overlay.Create(systemConfigurations)

	// Create CARAVELA's Remote Client
	caravelaCli := remote.NewClient(systemConfigurations.APIPort(), systemConfigurations.APITimeout())

	// Create Docker client
	dockerClient := docker.NewDockerClient(systemConfigurations)

	// Create the API server
	apiServer := rest.NewServer(systemConfigurations.APIPort())

	// Create a Caravela's Node passing all the external components and start its functions.
	thisNode := node.NewNode(systemConfigurations, overlayConfigured, caravelaCli, dockerClient, apiServer)

	err = thisNode.Start(join, joinIP)
	if err != nil {
		return err
	}

	return nil
}
