package cli

import (
	"github.com/strabox/caravela/api"
	"github.com/strabox/caravela/api/remote"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node"
)

func initNode(hostIP string, join bool, joinIP string) error {
	var systemConfigurations *configuration.Configuration
	var err error = nil

	// Create configuration structures from the configuration file (if it exists)
	if join {
		caravelaClient := remote.NewHttpClient(configuration.Default(hostIP, DockerEngineAPIVersion))

		systemConfigurations, err = caravelaClient.ObtainConfiguration(joinIP)
		if err != nil {
			return err
		}

		systemConfigurations, err = configuration.ObtainExternal(hostIP, DockerEngineAPIVersion, systemConfigurations)
		if err != nil {
			return err
		}
	} else {
		systemConfigurations, err = configuration.ReadFromFile(hostIP, DockerEngineAPIVersion)
		if err != nil {
			return err
		}
	}

	// Print/log the systemConfigurations values
	systemConfigurations.Print()

	// Create a CARAVELA node structure and start it functions
	thisNode := node.NewNode(systemConfigurations)
	err = thisNode.Start(join, joinIP)
	if err != nil {
		return err
	}

	// RunContainer CARAVELA REST API Server
	err = api.Initialize(systemConfigurations, thisNode)
	if err != nil {
		return err
	}

	return nil
}
