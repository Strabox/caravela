package main

import (
	"flag"
	"github.com/strabox/caravela/api"
	"github.com/strabox/caravela/api/client"
	"github.com/strabox/caravela/docker"
	"github.com/strabox/caravela/node"
	"github.com/strabox/caravela/node/configuration"
	"github.com/strabox/caravela/node/guid"
	"github.com/strabox/caravela/node/resources"
	"github.com/strabox/caravela/overlay"
	"github.com/strabox/caravela/overlay/chord"
	"log"
	"net"
)

// Flags used as input to CARAVELA container launch
var joinIP *string = flag.String("joinIP", "NOT_AN_IP", "Join a CARAVELA instance")
var hostIP *string = flag.String("hostIP", "NOT_AN_IP", "Docker Host IP")

func main() {
	flag.Parse() // Scan and parse the arguments list

	log.Println("##################################################################")
	log.Println("#          CARAVELA: A Cloud @ Edge                 000000       #")
	log.Println("#            Author: Andre Pires                  00000000000    #")
	log.Println("#  Email: pardal.pires@tecnico.ulisboa.pt           | ||| |      #")
	log.Println("#              IST/INESC-ID                        || ||| ||     #")
	log.Println("##################################################################")

	/*
		#################################################
		#	     Create Configurations Structure        #
		#################################################
	*/

	config := configuration.DefaultConfiguration(*hostIP)

	/*
		#################################################
		#	  Create and initializer Docker Client      #
		#################################################
	*/

	dockerClient := docker.NewDockerClient()
	dockerClient.Initialize("1.35") // TODO probably pass Docker API version it as an argument
	maxCPUs, maxRAM := dockerClient.GetDockerCPUandRAM()

	/*
		#################################################
		#   Create and initialize CARAVELA structures   #
		#################################################
	*/

	// Guid size initialization
	guid.InitializeGuid(config.ChordHashSizeBits)

	// Create Overlay struct (Chord overlay for my project)
	var overlay overlay.Overlay = chord.NewChordOverlay(guid.GuidSizeBytes(), *hostIP, config.OverlayPort,
		config.ChordVirtualNodes, config.ChordNumSuccessors, config.ChordTimeout)

	// Create CARAVELA's client
	var caravelaCli client.CaravelaClient = client.NewHttpClient(config.APIPort)

	// Node creation
	var thisNode *node.Node = node.NewNode(config, overlay, caravelaCli, *resources.NewResources(maxCPUs, maxRAM))

	/*
		#################################################
		#		     Create/Join an Overlay             #
		#################################################
	*/

	if net.ParseIP(*joinIP) != nil {
		overlay.Join(*joinIP, config.OverlayPort, thisNode)
	} else {
		overlay.Create(thisNode)
	}

	/*
		#################################################
		#    Start the CARAVELA's node functions        #
		#################################################
	*/

	thisNode.Start()

	/*
		#################################################
		#		 Start CARAVELA REST API Server         #
		#################################################
	*/

	api.Initialize(config.APIPort, thisNode)

}
