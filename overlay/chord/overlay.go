package chord

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/bluele/go-chord"
	nodeAPI "github.com/strabox/caravela/node/api"
	"github.com/strabox/caravela/overlay"
	"hash"
	"strconv"
	"strings"
	"time"
)

type Overlay struct {
	hashSizeBytes   int
	hostIP          string
	hostPort        int
	numVirtualNodes int
	numSuccessor    int
	timeout         time.Duration
	chordRing       *chord.Ring
}

func NewChordOverlay(hashSizeBytes int, hostIP string, hostPort int, numVnode int,
	numSuccessor int, timeout time.Duration) *Overlay {
	chordOverlay := &Overlay{}
	chordOverlay.hashSizeBytes = hashSizeBytes
	chordOverlay.hostIP = hostIP
	chordOverlay.hostPort = hostPort
	chordOverlay.numVirtualNodes = numVnode
	chordOverlay.numSuccessor = numSuccessor
	chordOverlay.timeout = timeout
	chordOverlay.chordRing = nil
	return chordOverlay
}

func (co *Overlay) init(thisNode nodeAPI.OverlayMembership) (*chord.Config, chord.Transport) {
	var hostname = co.hostIP + ":" + strconv.Itoa(co.hostPort)
	var chordListener = &Listener{thisNode}
	var config = chord.DefaultConfig(hostname)
	config.Delegate = chordListener
	config.NumVnodes = co.numVirtualNodes
	config.NumSuccessors = co.numSuccessor
	config.HashFunc = func() hash.Hash { return NewResourcesHash(co.hashSizeBytes, hostname) }

	log.Infoln("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
	log.Infoln("$                    CHORD OVERLAY CONFIGURATION                 $")
	log.Infof("$Hostname: %s                                       $", config.Hostname)
	log.Infof("$Num Virtual Nodes: %d                                            $", config.NumVnodes)
	log.Infof("$Num Successors: %d                                               $", config.NumSuccessors)
	log.Infoln("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")

	transport, err := chord.InitTCPTransport(fmt.Sprintf(":%d", co.hostPort), co.timeout)

	if err != nil {
		panic(fmt.Errorf("[Chord Overlay] Initializing Transport: %s", err))
	}

	return config, transport
}

/* ============================ Overlay Interface ============================ */

func (co *Overlay) Create(thisNode nodeAPI.OverlayMembership) {
	config, transport := co.init(thisNode)
	log.Debugln("[Chord Overlay] Creating a NEW CARAVELA instance ...")
	ring, err := chord.Create(config, transport)
	if err != nil {
		panic(fmt.Errorf("[Chord Overlay] Creating: %s", err))
	}
	co.chordRing = ring
	log.Debugln("[Chord Overlay] SUCCESS")
}

func (co *Overlay) Join(overlayNodeIP string, overlayNodePort int, thisNode nodeAPI.OverlayMembership) {
	config, transport := co.init(thisNode)
	log.Debugln("[Chord Overlay] Joining a CARAVELA instance ...")
	var joinHostname = overlayNodeIP + ":" + strconv.Itoa(overlayNodePort)
	ring, err := chord.Join(config, transport, joinHostname)
	if err != nil {
		panic(fmt.Errorf("[Chord Overlay] Joining: %s", err))
	}
	co.chordRing = ring
	log.Debugln("[Chord Overlay] SUCCESS")
}

func (co *Overlay) Lookup(key []byte) []*overlay.Node {
	if co.chordRing == nil {
		panic(fmt.Errorf("[Chord Overlay] Lookup failed. Chord not initialized"))
	}
	vnodes, _ := co.chordRing.Lookup(co.numSuccessor, key)
	res := make([]*overlay.Node, cap(vnodes))
	for index := range vnodes {
		res[index] = overlay.NewNode(strings.Split(vnodes[index].Host, ":")[0], vnodes[index].Id)
	}
	return res
}

func (co *Overlay) Leave() {
	if co.chordRing == nil {
		panic(fmt.Errorf("[Chord Overlay] Leave failed. Chord not initialized."))
	}
	co.chordRing.Leave()
}
