package cli

import (
	"github.com/strabox/caravela/api"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node"
)

func initNode(hostIP string, join bool, joinIP string) {
	// Create Configurations Structure
	config := configuration.Default(hostIP, DockerEngineAPIVersion)

	// Create NODE structure and start it functions
	thisNode := node.NewNode(config)
	thisNode.Start(join, joinIP)

	// Run CARAVELA REST API Server
	api.Initialize(config, thisNode)
}
