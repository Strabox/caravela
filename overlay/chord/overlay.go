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
	"math/big"
	"sync"
)

/*
Represents a Chord overlay local (for each node) structure.
*/
type Overlay struct {
	// Used to communicate interesting events to the application.
	appNode nodeAPI.OverlayMembership
	// Map of local virtual nodes IDs to the respective predecessors IDs
	predecessors sync.Map
	// Number of virtual nodes up and running
	virtualNodesRunning int
	// Physical node ID (Higher ID of all the virtual nodes)
	localID []byte

	// ============== CHORD Related Parameters ==================
	// Size of the hash (in bytes) produced by the lookup hash function.
	hashSizeBytes int
	// IP address of the local node.
	hostIP string
	// Port where local node is running the chord overlay daemon.
	hostPort int
	// Number of virtual nodes in the local "physical node".
	numVirtualNodes int
	// Number of successor nodes maintained by the chord.
	numSuccessors int
	// Timeout for chord overlay messages (pings, etc).
	timeout time.Duration
	// Chord ring structure from the library used (github.com/strabox/go-chord).
	chordRing *chord.Ring
}

/*
Create a new Chord overlay structure.
*/
func NewChordOverlay(hashSizeBytes int, hostIP string, hostPort int,
	numVirtualNodes int, numSuccessors int, timeout time.Duration) *Overlay {
	chordOverlay := &Overlay{
		appNode:             nil,
		predecessors:        sync.Map{},
		virtualNodesRunning: 0,
		localID:             nil,

		hashSizeBytes:   hashSizeBytes,
		hostIP:          hostIP,
		hostPort:        hostPort,
		numVirtualNodes: numVirtualNodes,
		numSuccessors:   numSuccessors,
		timeout:         timeout,
		chordRing:       nil,
	}
	return chordOverlay
}

/*
Initialize the chord overlay and its respective inner structures.
*/
func (co *Overlay) initialize(appNode nodeAPI.OverlayMembership) (*chord.Config, chord.Transport) {
	co.appNode = appNode
	hostname := co.hostIP + ":" + strconv.Itoa(co.hostPort)
	chordListener := &Listener{chordOverlay: co}
	config := chord.DefaultConfig(hostname)

	config.Delegate = chordListener
	config.NumVnodes = co.numVirtualNodes
	config.NumSuccessors = co.numSuccessors
	config.HashFunc = func() hash.Hash { return NewResourcesHash(co.hashSizeBytes, hostname) }

	log.Debug("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
	log.Debug("$                    CHORD OVERLAY CONFIGURATION                 $")
	log.Debugf("$Hostname: %s                                      $", config.Hostname)
	log.Debugf("$Num Virtual Nodes: %d                                            $", config.NumVnodes)
	log.Debugf("$Num Successors: %d                                               $", config.NumSuccessors)
	log.Debug("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")

	// Initialize the TCP transport stack used in this chord implementation
	transport, err := chord.InitTCPTransport(fmt.Sprintf(":%d", co.hostPort), co.timeout)

	if err != nil {
		log.Fatalf(util.LogTag("[Chord]")+"Initialize transport stack error: %s", err)
	}

	return config, transport
}

/*
Called when a new virtual node of this physical node has joined the chord ring.
 */
func (co *Overlay) newLocalVirtualNode(localVirtualNodeID []byte, predecessorNode *overlay.Node) {
	newLocalVirtualNodeID := big.NewInt(0)
	newLocalVirtualNodeID.SetBytes(localVirtualNodeID)
	if co.localID == nil {
		co.localID = localVirtualNodeID
	} else {
		localID := big.NewInt(0)
		localID.SetBytes(co.localID)
		if newLocalVirtualNodeID.Cmp(localID) >= 0 {
			co.localID = localVirtualNodeID
		}
	}
	co.virtualNodesRunning++
	co.predecessors.Store(newLocalVirtualNodeID.String(), predecessorNode)
	co.appNode.AddTrader(localVirtualNodeID) // Alert the node for the new virtual node/trader
}

/*
Called when the predecessor of a virtual node of the physical node changes.
e.g. Due to a crash in the previous predecessor or because he left the chord ring.
 */
func (co *Overlay) predecessorNodeChanged(localVirtualNodeID []byte, predecessorNode *overlay.Node) {
	vNodeID := big.NewInt(0)
	vNodeID.SetBytes(localVirtualNodeID)
	co.predecessors.Store(vNodeID.String(), predecessorNode)
}

/*
Verify if the chord ring is initialized, if it is not it starts panicking because
nothing can be done without the chord overlay working.
 */
func (co *Overlay) initialized() {
	if co.chordRing == nil {
		panic(fmt.Errorf(util.LogTag("[Chord]") + "Lookup failed. Chord not initialized"))
	}
}

/* ============================ Overlay Interface ============================ */

func (co *Overlay) Create(appNode nodeAPI.OverlayMembership) {
	config, transport := co.initialize(appNode)
	log.Debugln(util.LogTag("[Chord]") + "Creating a NEW CARAVELA instance ...")
	ring, err := chord.Create(config, transport)
	if err != nil {
		panic(fmt.Errorf(util.LogTag("[Chord]")+"Create error: %s", err))
	}
	co.chordRing = ring
	log.Debugln(util.LogTag("[Chord]") + "Create SUCCESS")
}

func (co *Overlay) Join(overlayNodeIP string, overlayNodePort int, appNode nodeAPI.OverlayMembership) {
	config, transport := co.initialize(appNode)
	log.Debug(util.LogTag("[Chord]") + "Joining a CARAVELA instance ...")
	var joinHostname = overlayNodeIP + ":" + strconv.Itoa(overlayNodePort)
	ring, err := chord.Join(config, transport, joinHostname)
	if err != nil {
		panic(fmt.Errorf(util.LogTag("[Chord]")+"Join error: %s", err))
	}
	co.chordRing = ring
	log.Debug(util.LogTag("[Chord]") + "Join SUCCESS")
}

func (co *Overlay) Lookup(key []byte) []*overlay.Node {
	co.initialized()

	virtualNodes, err := co.chordRing.Lookup(co.numSuccessors, key)
	if err != nil {
		log.Errorf(util.LogTag("[Chord]")+"Lookup error: %s", err)
	}
	res := make([]*overlay.Node, len(virtualNodes))
	for index := range virtualNodes {
		res[index] = overlay.NewNode(strings.Split(virtualNodes[index].Host, ":")[0], virtualNodes[index].Id)
	}
	return res
}

func (co *Overlay) Neighbors(nodeID []byte) []*overlay.Node {
	co.initialized()

	id := big.NewInt(0)
	id.SetBytes(nodeID)
	log.Debugf(util.LogTag("[Chord]")+"Node %s", id.String())
	res := make([]*overlay.Node, 0)
	nodes := co.Lookup(nodeID)
	if len(nodes) > 1 {
		hmm := big.NewInt(0)
		hmm.SetBytes(nodes[1].GUID())
		log.Debugf(util.LogTag("[Chord]")+"Successor %s", hmm.String())
		res = append(res, nodes[1]) // The successor of the given node
	}
	predecessorNode, exist := co.predecessors.Load(id.String())
	if exist {
		node, ok := predecessorNode.(*overlay.Node)
		if ok {
			hmm := big.NewInt(0)
			hmm.SetBytes(node.GUID())
			log.Debugf(util.LogTag("[Chord]")+"Predecessor %s", hmm.String())
			res = append(res, node) // The predecessor of the given node
		}
	}
	return res
}

func (co *Overlay) NodeID() []byte {
	co.initialized()

	if co.localID != nil && co.virtualNodesRunning == co.numVirtualNodes {
		hmm := big.NewInt(0)
		hmm.SetBytes(co.localID)
		log.Debugf(util.LogTag("[Chord]")+"Node ID %s", hmm.String())
		return co.localID
	} else {
		log.Debugf(util.LogTag("[Chord]") + "Node ID not fixed yes!!")
		return nil
	}
}

func (co *Overlay) Leave() {
	co.initialized()

	err := co.chordRing.Leave()
	if err != nil {
		log.Errorf(util.LogTag("[Chord]")+"Leave error: %s", err)
	}
}
