package main

import (
	"flag"
	"fmt"
	"github.com/strabox/caravela/api"
	"github.com/strabox/caravela/api/client"
	"github.com/strabox/caravela/docker"
	"github.com/strabox/caravela/overlay/chordoverlay"
	"github.com/strabox/caravela/overlay"
	"github.com/strabox/caravela/node/resources"
	"github.com/strabox/caravela/node"
	"github.com/strabox/caravela/node/guid"
	"github.com/strabox/caravela/node/configuration"
	"net"
)

// Flags used as input to CARAVELA container launch
var joinIP *string = flag.String("joinIP", "NOT_AN_IP", "Join a CARAVELA instance")
var hostIP *string = flag.String("hostIP", "NOT_AN_IP", "Docker Host IP")

func main() {
	flag.Parse() // Scan and parse the arguments list

	fmt.Println("##################################################################")
	fmt.Println("#          CARAVELA: A Cloud @ Edge                 000000       #")
	fmt.Println("#            Author: Andre Pires                  00000000000    #")
	fmt.Println("#  Email: pardal.pires@tecnico.ulisboa.pt           | ||| |      #")
	fmt.Println("#              IST/INESC-ID                        || ||| ||     #")
	fmt.Println("##################################################################")
	
	/*
		#################################################
		#	           Create Configurations            #
		#################################################
	*/
	
	config := configuration.DefaultConfiguration(*hostIP)
	
	/*
		#################################################
		#	      Initializing Docker Client            #
		#################################################
	*/

	docker.Initialize("1.35") 	// TODO probably pass Docker API version it as an argument
	maxCPUs, maxRAM := docker.GetDockerCPUandRAM()
	
	/*
		#################################################
		#   Create and initialize CARAVELA structures   #
		#################################################
	*/
	
	// Guid size initialization
	guid.InitializeGuid(config.ChordHashSizeBits)
	
	// Overlay initialization (Chord overlay for my project)
	var overlay overlay.Overlay = chordoverlay.NewChordOverlay(guid.GuidSizeBytes(), *hostIP, config.OverlayPort, 
		config.ChordVirtualNodes, config.ChordNumSuccessors, config.ChordTimeout)
	
	//
	var caravelaCli client.CaravelaClient = client.NewHttpClient()
	
	// Resources Mapping creation
	var resourcesMap *resources.ResourcesMap = resources.NewResourcesMap(config.CpuPartitions, config.RamPartitions)
	resourcesMap.Print()

	// Node creation
	var thisNode *node.Node = node.NewNode(config, overlay, caravelaCli, resourcesMap, config.ChordVirtualNodes, *resources.NewResources(maxCPUs, maxRAM))
	
	/*
		#################################################
		#		   Initializing Overlay (CHORD)         #
		#################################################
	*/
	
	if net.ParseIP(*joinIP) != nil {
		overlay.Join(*joinIP, config.OverlayPort, thisNode)	
	} else {
		overlay.Create(thisNode)
	}

	/*
		#################################################
		#		Initializing CARAVELA REST API          #
		#################################################
	*/

	api.Initialize(config.APIPort, thisNode)

}
