package chordoverlay

import (
	"fmt"
	"github.com/bluele/go-chord"
	"github.com/strabox/caravela/node"
	"net"
	"strconv"
	"time"
)

type ChordOverlay struct {
	hostIP       string
	hostPort     int
	numVnode     int
	numSuccessor int
	timeoutMili  int64
	chordRing    *chord.Ring
}

func NewChordOverlay(hostIP string, hostPort int, numVnode int, numSuccessor int, timeoutMili int64) *ChordOverlay {
	chordOverlay := &ChordOverlay{}
	chordOverlay.hostIP = hostIP
	chordOverlay.hostPort = hostPort
	chordOverlay.numVnode = numVnode
	chordOverlay.numSuccessor = numSuccessor
	chordOverlay.timeoutMili = timeoutMili
	return chordOverlay
}

func (co *ChordOverlay) init() (*chord.Config, chord.Transport) {
	var hostname = co.hostIP + ":" + strconv.Itoa(co.hostPort)
	var chordListner = &ChordListner{}
	var config = chord.DefaultConfig(hostname)
	config.Delegate = chordListner
	config.NumVnodes = co.numVnode
	config.NumSuccessors = co.numSuccessor

	fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
	fmt.Println("$                    CHORD OVERLAY CONFIGURATION                 $")
	fmt.Printf("$Hostname: %s                                       $\n", config.Hostname)
	fmt.Printf("$Num Virtual Nodes: %d                                            $\n", config.NumVnodes)
	fmt.Printf("$Num Successors: %d                                               $\n", config.NumSuccessors)
	fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")

	transport, err := chord.InitTCPTransport(":"+strconv.Itoa(co.hostPort), time.Duration(co.timeoutMili)*time.Millisecond)

	if err != nil {
		panic(fmt.Errorf("[Chord Overlay] Initializing Transport: %s", err))
	}

	return config, transport
}

/* ============================ Overlay Interface ============================ */

func (co *ChordOverlay) Create() {
	config, transport := co.init()
	fmt.Println("[Chord Overlay] Creating a NEW CARAVELA instance ...")
	ring, err := chord.Create(config, transport)
	if err != nil {
		panic(fmt.Errorf("[Chord Overlay] Creating: %s", err))
	}
	co.chordRing = ring
	fmt.Println("[Chord Overlay] SUCCESS")
}

func (co *ChordOverlay) Join(overlayNodeIP string, overlayNodePort int) {
	config, transport := co.init()
	fmt.Println("[Chord Overlay] Joining a CARAVELA instance ...")
	var joinHostname = overlayNodeIP + ":" + strconv.Itoa(overlayNodePort)
	ring, err := chord.Join(config, transport, joinHostname)
	if err != nil {
		panic(fmt.Errorf("[Chord Overlay] Joining: %s", err))
	}
	co.chordRing = ring
	fmt.Println("[Chord Overlay] SUCCESS")
}

func (co *ChordOverlay) Lookup(resources node.Resources) []*node.RemoteNode {
	//TODO
	return nil
}

func (co *ChordOverlay) Leave() {
	co.chordRing.Leave()
}

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
	fmt.Println("$                    CHORD OVERLAY CONFIGURATION                 $")
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
