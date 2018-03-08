package chordoverlay

import (
	"fmt"
	"github.com/bluele/go-chord"
	"github.com/strabox/caravela/node/guid"
	"github.com/strabox/caravela/node/local"
	"github.com/strabox/caravela/overlay"
	"strconv"
	"strings"
	"hash"
	"time"
)

type ChordOverlay struct {
	hashSizeBytes int
	hostIP       string
	hostPort     int
	numVnode     int
	numSuccessor int
	timeout      time.Duration
	chordRing    *chord.Ring
}

const NUM_NODES_IN_LOOKUP = 3

func NewChordOverlay(hashSizeBytes int, hostIP string, hostPort int, numVnode int,
		numSuccessor int, timeout time.Duration) *ChordOverlay {
	chordOverlay := &ChordOverlay{}
	chordOverlay.hashSizeBytes = hashSizeBytes
	chordOverlay.hostIP = hostIP
	chordOverlay.hostPort = hostPort
	chordOverlay.numVnode = numVnode
	chordOverlay.numSuccessor = numSuccessor
	chordOverlay.timeout = timeout
	chordOverlay.chordRing = nil
	return chordOverlay
}

func (co *ChordOverlay) init(thisNode local.LocalNode) (*chord.Config, chord.Transport) {
	var hostname = co.hostIP + ":" + strconv.Itoa(co.hostPort)
	var chordListner = &ChordListner{thisNode}
	var config = chord.DefaultConfig(hostname)
	config.Delegate = chordListner
	config.NumVnodes = co.numVnode
	config.NumSuccessors = co.numSuccessor
	config.HashFunc = func() hash.Hash { return NewResourcesHash(co.hashSizeBytes, hostname) }
	
	fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
	fmt.Println("$                    CHORD OVERLAY CONFIGURATION                 $")
	fmt.Printf("$Hostname: %s                                       $\n", config.Hostname)
	fmt.Printf("$Num Virtual Nodes: %d                                            $\n", config.NumVnodes)
	fmt.Printf("$Num Successors: %d                                               $\n", config.NumSuccessors)
	fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")

	transport, err := chord.InitTCPTransport(fmt.Sprintf(":%d",co.hostPort), co.timeout)

	if err != nil {
		panic(fmt.Errorf("[Chord Overlay] Initializing Transport: %s", err))
	}

	return config, transport
}

/* ============================ Overlay Interface ============================ */

func (co *ChordOverlay) Create(thisNode local.LocalNode) {
	config, transport := co.init(thisNode)
	fmt.Println("[Chord Overlay] Creating a NEW CARAVELA instance ...")
	ring, err := chord.Create(config, transport)
	if err != nil {
		panic(fmt.Errorf("[Chord Overlay] Creating: %s", err))
	}
	co.chordRing = ring
	fmt.Println("[Chord Overlay] SUCCESS")
}

func (co *ChordOverlay) Join(overlayNodeIP string, overlayNodePort int, thisNode local.LocalNode) {
	config, transport := co.init(thisNode)
	fmt.Println("[Chord Overlay] Joining a CARAVELA instance ...")
	var joinHostname = overlayNodeIP + ":" + strconv.Itoa(overlayNodePort)
	ring, err := chord.Join(config, transport, joinHostname)
	if err != nil {
		panic(fmt.Errorf("[Chord Overlay] Joining: %s", err))
	}
	co.chordRing = ring
	fmt.Println("[Chord Overlay] SUCCESS")
}

func (co *ChordOverlay) Lookup(key guid.Guid) []*overlay.RemoteNode {
	if co.chordRing == nil {
		panic(fmt.Errorf("[Chord Overlay] Lookup failed. Chord not initialized."))	
	}
	vnodes, _ := co.chordRing.Lookup(NUM_NODES_IN_LOOKUP, key.Bytes())
	res := make([]*overlay.RemoteNode, cap(vnodes))
	for index, _ := range vnodes {
		res[index] = overlay.NewRemoteNode(strings.Split(vnodes[index].Host, ":")[0], guid.NewGuidBytes(vnodes[index].Id))
	}
	return res
}

func (co *ChordOverlay) Leave() {
	if co.chordRing == nil {
		panic(fmt.Errorf("[Chord Overlay] Leave failed. Chord not initialized."))	
	}
	co.chordRing.Leave()
}

