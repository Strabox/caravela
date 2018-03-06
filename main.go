package main

import (
	"flag"
	"fmt"
	"github.com/strabox/caravela/api"
	"github.com/strabox/caravela/docker"
	"github.com/strabox/caravela/overlay/chordoverlay"
	"github.com/strabox/caravela/overlay"
	"github.com/strabox/caravela/node/resources"
	"github.com/strabox/caravela/node"
	"github.com/strabox/caravela/node/guid"
	"net"
)

const CARAVELA_PORT = 8000

const CHORD_PORT = 8001
const CHORD_TIMEOUT_MILIS = 2000
const CHORD_V_NODES = 8
const CHORD_NUM_SUCCESSORS = 3
const CHORD_HASH_SIZE_BITS = 160

var CPU_PARTITIONS []resources.ResourcePerc = []resources.ResourcePerc{resources.ResourcePerc{1,50}, resources.ResourcePerc{2,50}}
var RAM_PARTITIONS []resources.ResourcePerc = []resources.ResourcePerc{resources.ResourcePerc{256,50}, resources.ResourcePerc{512,50}}

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
		#	      Initializing Docker Client            #
		#################################################
	*/

	docker.Initialize("1.35") 	// TODO probably pass Docker API version it as an argument
	cpu, ram := docker.GetDockerCPUandRAM()
	
	/*
		#################################################
		#   Create and initialize CARAVELA structures   #
		#################################################
	*/
	
	// Guid size initialization
	guid.InitializeGuid(CHORD_HASH_SIZE_BITS)
	
	// Overlay initialization (Chord overlay for my project)
	var overlay overlay.Overlay = chordoverlay.NewChordOverlay(guid.GuidSizeBytes(), *hostIP, CHORD_PORT, CHORD_V_NODES, CHORD_NUM_SUCCESSORS, CHORD_TIMEOUT_MILIS)
	
	// Resources Mapping creation
	var resourcesMap *resources.ResourcesMap = resources.NewResourcesMap(CPU_PARTITIONS, RAM_PARTITIONS)
	resourcesMap.Print()

	// Node creation
	var thisNode *node.Node = node.NewNode(overlay, resourcesMap, CHORD_V_NODES)
	var supplier *node.Supplier = node.NewSupplier(thisNode, *resources.NewResources(cpu, ram))
	thisNode.SetSupplier(supplier)
	
	
	/*
		#################################################
		#		   Initializing Overlay CHORD           #
		#################################################
	*/
	
	if net.ParseIP(*joinIP) != nil {
		thisNode.Overlay().Join(*joinIP, CHORD_PORT, thisNode)	
	} else {
		thisNode.Overlay().Create(thisNode)
	}

	/*
		#################################################
		#		Initializing CARAVELA REST API          #
		#################################################
	*/

	api.Initialize(CARAVELA_PORT, thisNode)

}
