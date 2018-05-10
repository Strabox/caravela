package chord

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	nodeAPI "github.com/strabox/caravela/node/api"
	"github.com/strabox/caravela/overlay"
	"github.com/strabox/caravela/util"
	"github.com/strabox/go-chord"
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

	log.Debug("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
	log.Debug("$                    CHORD OVERLAY CONFIGURATION                 $")
	log.Debugf("$Hostname: %s                                      $", config.Hostname)
	log.Debugf("$Num Virtual Nodes: %d                                            $", config.NumVnodes)
	log.Debugf("$Num Successors: %d                                               $", config.NumSuccessors)
	log.Debug("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")

	transport, err := chord.InitTCPTransport(fmt.Sprintf(":%d", co.hostPort), co.timeout)

	if err != nil {
		panic(fmt.Errorf(util.LogTag("[Chord]")+"Initializing Transport: %s", err))
	}

	return config, transport
}

/* ============================ Overlay Interface ============================ */

func (co *Overlay) Create(thisNode nodeAPI.OverlayMembership) {
	config, transport := co.init(thisNode)
	log.Debugln(util.LogTag("[Chord]") + "Creating a NEW CARAVELA instance ...")
	ring, err := chord.Create(config, transport)
	if err != nil {
		panic(fmt.Errorf(util.LogTag("[Chord]")+"Creating: %s", err))
	}
	co.chordRing = ring
	log.Debugln(util.LogTag("[Chord]") + "SUCCESS")
}

func (co *Overlay) Join(overlayNodeIP string, overlayNodePort int, thisNode nodeAPI.OverlayMembership) {
	config, transport := co.init(thisNode)
	log.Debug(util.LogTag("[Chord]") + "Joining a CARAVELA instance ...")
	var joinHostname = overlayNodeIP + ":" + strconv.Itoa(overlayNodePort)
	ring, err := chord.Join(config, transport, joinHostname)
	if err != nil {
		panic(fmt.Errorf(util.LogTag("[Chord]")+"Joining: %s", err))
	}
	co.chordRing = ring
	log.Debug(util.LogTag("[Chord]") + "SUCCESS")
}

func (co *Overlay) Lookup(key []byte) []*overlay.Node {
	if co.chordRing == nil {
		panic(fmt.Errorf(util.LogTag("[Chord]") + "Lookup failed. Chord not initialized"))
	}
	virtualNodes, _ := co.chordRing.Lookup(co.numSuccessor, key)
	res := make([]*overlay.Node, cap(virtualNodes))
	for index := range virtualNodes {
		res[index] = overlay.NewNode(strings.Split(virtualNodes[index].Host, ":")[0], virtualNodes[index].Id)
	}
	return res
}

func (co *Overlay) Leave() {
	if co.chordRing == nil {
		panic(fmt.Errorf(util.LogTag("[Chord]") + "Leave failed. Chord not initialized."))
	}
	co.chordRing.Leave()
}
