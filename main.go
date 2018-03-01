package main

import (
	"flag"
	"fmt"
	"github.com/strabox/caravela/api"
	"github.com/strabox/caravela/docker"
	"github.com/strabox/caravela/overlay"
)

const CARAVELA_PORT = 8000

const CHORD_PORT = 8001
const CHORD_TIMEOUT_MILIS = 2000
const CHORD_V_NODES = 3
const CHORD_NUM_SUCCESSORS = 3

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

	docker.Initialize("1.35") // TODO probably pass Docker API version it as an argument

	/*
		#################################################
		#		   Initializing Overlay CHORD           #
		#################################################
	*/

	overlay.Initialize(*hostIP, *joinIP, CHORD_PORT, CHORD_V_NODES, CHORD_NUM_SUCCESSORS, CHORD_TIMEOUT_MILIS)

	/*
		#################################################
		#		Initializing CARAVELA REST API          #
		#################################################
	*/

	api.Initialize(CARAVELA_PORT)

}
