package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/gorilla/mux"
	"github.com/strabox/caravela/api/discovery"
	"github.com/strabox/caravela/membership"
	"github.com/strabox/caravela/node"
	"github.com/strabox/go-chord"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

const CARAVELA_PORT = 8000

const CHORD_PORT = 8001
const CHORD_TIMEOUT = 2 * time.Second
const CHORD_V_NODES = 3
const CHORD_NUM_SUCCESSORS = 3

// Flags used as input to CARAVELA
var joinIP *string = flag.String("joinIP", "NOT_AN_IP", "Join a CARAVELA instance")
var hostIP *string = flag.String("hostIP", "NOT_AN_IP", "Docker Host IP")

func main() {
	var dockerClient *client.Client
	flag.Parse() // Scan and parse the arguments list

	fmt.Println("##################################################################")
	fmt.Println("#          CARAVELA: A Cloud @ Edge                 000000       #")
	fmt.Println("#            Author: Andre Pires                   000000000     #")
	fmt.Println("#  Email: pardal.pires@tecnico.ulisboa.pt           | | | |      #")
	fmt.Println("#              IST/INESC-ID                        |  | |  |     #")
	fmt.Println("##################################################################")

	/*
		#################################################
		#	  Initializing Docker Client                #
		#################################################
	*/

	dockerClient = initializeDockerClient("1.35")

	// AKA: TRY ZONZE
	cpu, ram := getDockerCPUandRAM(dockerClient)
	fmt.Printf("%d %d\n", cpu, ram)

	one := big.NewInt(1)
	test1 := big.NewInt(1)
	test2 := big.NewInt(2)
	test3 := big.NewInt(160)
	test1.Exp(test2, test3, nil)
	test1.Sub(test1, one)
	fmt.Println("Test: ", test1.Bytes())

	var g = node.NewGuid(node.GUID_BYTES_SIZE)
	g.PrintDecimal()
	var h = node.NewResourcesHash(node.GUID_BYTES_SIZE)
	h.Write(g.GetKey())
	fmt.Println(h.Sum(nil))
	fmt.Println("ENDA")
	return

	/*
		#################################################
		#			  Initializing CHORD                #
		#################################################
	*/
	if net.ParseIP(*hostIP) == nil {
		fmt.Println("ERR Invalid Host IP")
		os.Exit(1)
	}

	var hostname = *hostIP + ":" + strconv.Itoa(CHORD_PORT)
	var chordListner = &membership.ChordListner{}
	var config = chord.DefaultConfig(hostname)

	config.Delegate = chordListner
	config.NumVnodes = CHORD_V_NODES
	config.NumSuccessors = CHORD_NUM_SUCCESSORS
	fmt.Printf("Chord configuration->\nHostname: %s \nNum Virtual Nodes: %d \nNum Sucessors: %d \n", config.Hostname, config.NumVnodes, config.NumSuccessors)

	var err error
	var transport chord.Transport
	transport, err = chord.InitTCPTransport(":"+strconv.Itoa(CHORD_PORT), CHORD_TIMEOUT)
	var Ring *chord.Ring

	if err != nil {
		fmt.Println("ERR Init Transport: ", err)
		os.Exit(1)
	}

	if net.ParseIP(*joinIP) != nil {
		fmt.Println("Joining a CARAVELA instance ...")
		var joinHostname = *joinIP + ":" + strconv.Itoa(CHORD_PORT)
		Ring, err = chord.Join(config, transport, joinHostname)
		if err != nil {
			fmt.Println("ERR Joining: ", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("Creating a NEW CARAVELA instance ...")
		Ring, err = chord.Create(config, transport)

		if err != nil {
			fmt.Println("ERR Creating: ", err)
			os.Exit(1)
		}
	}

	// Represents the Ring where the host is inserted on.
	discovery.Ring = Ring
	fmt.Println("Ring: ", Ring)

	/*
		#################################################
		#		Initializing CARAVELA REST API          #
		#################################################
	*/
	fmt.Println("Initializing CARAVELA REST API ...")
	router := mux.NewRouter()
	router.HandleFunc("/debug/status", discovery.ChordStatus).Methods("GET")
	router.HandleFunc("/lookup/{key}", discovery.ChordLookup).Methods("GET")
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(CARAVELA_PORT), router))
	fmt.Println("CARAVELA UP AND RUNNING ...")
}

/*
initializeDockerClient creates and initialize a docker client
*/
func initializeDockerClient(runningDockerVersion string) *client.Client {
	cli, err := client.NewEnvClient()

	if err != nil {
		fmt.Println("[Creating Docker SDK Client] ", err)
		panic(err)
	}

	cli, err = client.NewClientWithOpts(client.WithVersion(runningDockerVersion))
	if err != nil {
		fmt.Println("[Init Docker SDK Client] ", err)
		panic(err)
	}

	return cli
}

/*
getDockerCPUandRAM get CPU and RAM dedicated to Docker engine (BY the user)
*/
func getDockerCPUandRAM(client *client.Client) (uint, uint) {
	ctx := context.Background()
	info, _ := client.Info(ctx)
	cpu := uint(info.NCPU)
	ram := uint(info.MemTotal / 1000000) //Return in MB
	return cpu, ram
}
