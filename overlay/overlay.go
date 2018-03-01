package overlay

import (
	"fmt"
	"github.com/bluele/go-chord"
	"net"
	"strconv"
	"time"
)

var Overlay *chord.Ring = nil

func Initialize(hostIP string, joinIP string, overlayPort int, numVnodes int, numSuccessors int, overlayTimeoutMili int64) {

	if net.ParseIP(hostIP) == nil {
		panic(fmt.Errorf("[Overlay] Invalid Host IP: %s", hostIP))
	}

	var hostname = hostIP + ":" + strconv.Itoa(overlayPort)
	var chordListner = &ChordListner{}
	var config = chord.DefaultConfig(hostname)
	config.Delegate = chordListner
	config.NumVnodes = numVnodes
	config.NumSuccessors = numSuccessors

	fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
	fmt.Println("$                      OVERLAY CONFIGURATION                     $")
	fmt.Printf("$Hostname: %s                                       $\n", config.Hostname)
	fmt.Printf("$Num Virtual Nodes: %d                                            $\n", config.NumVnodes)
	fmt.Printf("$Num Successors: %d                                               $\n", config.NumSuccessors)
	fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")

	var err error
	var transport chord.Transport
	transport, err = chord.InitTCPTransport(":"+strconv.Itoa(overlayPort), time.Duration(overlayTimeoutMili)*time.Millisecond)
	var ring *chord.Ring = nil

	if err != nil {
		panic(fmt.Errorf("[Overlay] Initializing Transport: %s", err))
	}

	if net.ParseIP(joinIP) != nil {
		fmt.Println("[Overlay] Joining a CARAVELA instance ...")
		var joinHostname = joinIP + ":" + strconv.Itoa(overlayPort)
		ring, err = chord.Join(config, transport, joinHostname)
		if err != nil {
			panic(fmt.Errorf("[Overlay] Joining: %s", err))
		}
	} else {
		fmt.Println("[Overlay] Creating a NEW CARAVELA instance ...")
		ring, err = chord.Create(config, transport)

		if err != nil {
			panic(fmt.Errorf("[Overlay] Creating: %s", err))
		}
	}

	// Represents the Ring where the host is inserted on.
	Overlay = ring
	fmt.Println("[Overlay] Initialized Success")
}
