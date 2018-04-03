package cli

import (
	"github.com/strabox/caravela/api"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node"
)

func initNode(hostIP string, join bool, joinIP string) {

	/*
		#################################################
		#	     Create Configurations Structure        #
		#################################################
	*/

	config := configuration.DefaultConfiguration(hostIP, "1.35") // TODO: probably pass Docker API version it as an argument

	/*
		#################################################
		#  Create NODE structure and start it functions #
		#################################################
	*/

	thisNode := node.NewNode(config)

	thisNode.Start(join, joinIP)

	/*
		#################################################
		#		 Start CARAVELA REST API Server         #
		#################################################
	*/

	api.Initialize(config, thisNode)
}
