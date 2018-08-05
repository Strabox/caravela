package chord

import (
	"context"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/overlay/types"
	"github.com/strabox/caravela/util"
	"github.com/strabox/go-chord"
	"hash"
	"math/big"
	"strconv"
	"sync"
	"time"
)

// Represents a Chord overlay local (for each node) structure.
type Chord struct {
	// Used to communicate interesting events to the application.
	appNode types.OverlayMembership
	// Map of local virtual nodes IDs to the respective predecessors IDs
	predecessors sync.Map
	// Number of virtual nodes up and running
	virtualNodesRunning int
	// Physical node ID (Higher ID of all the virtual nodes)
	localID []byte

	// ============== CHORD Related Fields  ==================
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

// Create a new Chord overlay structure.
func New(hashSizeBytes int, hostIP string, hostPort int,
	numVirtualNodes int, numSuccessors int, timeout time.Duration) *Chord {

	chordOverlay := &Chord{
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

// Initialize the chord overlay and its respective inner structures.
func (co *Chord) initialize(appNode types.OverlayMembership) (*chord.Config, chord.Transport, error) {
	co.appNode = appNode
	hostname := co.hostIP + ":" + strconv.Itoa(co.hostPort)
	chordListener := &Listener{chordOverlay: co}
	config := chord.DefaultConfig(hostname)

	config.Delegate = chordListener
	config.NumVnodes = co.numVirtualNodes
	config.NumSuccessors = co.numSuccessors
	config.HashFunc = func() hash.Hash { return NewResourcesHash(co.hashSizeBytes, hostname) }

	// Initialize the TCP transport stack used in the chord implementation
	transport, err := chord.InitTCPTransport(fmt.Sprintf(":%d", co.hostPort), co.timeout)

	if err != nil {
		return nil, nil, err
	}

	return config, transport, nil
}

// Called when a new virtual node of this physical node has joined the chord ring.
func (co *Chord) newLocalVirtualNode(localVirtualNodeID []byte, predecessorNode *types.OverlayNode) {
	if co.virtualNodesRunning == co.numVirtualNodes {
		return
	}

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

// Called when the predecessor of a virtual node of the physical node changes.
// e.g. Due to a crash in the previous predecessor or because he left the chord ring.
func (co *Chord) predecessorNodeChanged(localVirtualNodeID []byte, predecessorNode *types.OverlayNode) {
	vNodeID := big.NewInt(0)
	vNodeID.SetBytes(localVirtualNodeID)
	co.predecessors.Store(vNodeID.String(), predecessorNode)
}

/* ============================ Overlay Interface ============================ */

func (co *Chord) Create(ctx context.Context, appNode types.OverlayMembership) error {
	config, transport, err := co.initialize(appNode)
	if err != nil {
		return err
	}

	ring, err := chord.Create(config, transport)
	if err != nil {
		return fmt.Errorf("create chord error: %s", err)
	}

	co.chordRing = ring
	return nil
}

func (co *Chord) Join(ctx context.Context, overlayNodeIP string, overlayNodePort int, appNode types.OverlayMembership) error {
	config, transport, err := co.initialize(appNode)
	if err != nil {
		return err
	}

	var joinHostname = overlayNodeIP + ":" + strconv.Itoa(overlayNodePort)
	ring, err := chord.Join(config, transport, joinHostname)
	if err != nil {
		return fmt.Errorf("join chord error: %s", err)
	}
	co.chordRing = ring
	return nil
}

func (co *Chord) Lookup(ctx context.Context, key []byte) ([]*types.OverlayNode, error) {
	virtualNodes, err := co.chordRing.Lookup(co.numSuccessors, key)
	if err != nil {
		log.Errorf(util.LogTag("Chord")+"Lookup error: %s", err)
		return make([]*types.OverlayNode, 0), fmt.Errorf("lookup error")
	}

	res := make([]*types.OverlayNode, len(virtualNodes))
	for index := range virtualNodes {
		nodeIP, nodePort := util.ObtainIpPort(virtualNodes[index].Host)
		res[index] = types.NewOverlayNode(nodeIP, nodePort, virtualNodes[index].Id)
	}
	return res, nil
}

func (co *Chord) Neighbors(ctx context.Context, nodeID []byte) ([]*types.OverlayNode, error) {
	id := big.NewInt(0)
	id.SetBytes(nodeID)

	res := make([]*types.OverlayNode, 0)
	nodes, err := co.Lookup(ctx, nodeID) // TODO: Optional, avoid this lookup.
	if err != nil {
		return make([]*types.OverlayNode, 0), err
	}

	if len(nodes) > 1 {
		hmm := big.NewInt(0)
		hmm.SetBytes(nodes[1].GUID())
		res = append(res, nodes[1]) // The successor of the given node
	}
	predecessorNode, exist := co.predecessors.Load(id.String())
	if exist {
		node, ok := predecessorNode.(*types.OverlayNode)
		if ok {
			hmm := big.NewInt(0)
			hmm.SetBytes(node.GUID())
			res = append(res, node) // The predecessor of the given node
		}
	}
	return res, nil
}

func (co *Chord) NodeID(ctx context.Context) ([]byte, error) {
	if co.localID != nil && co.virtualNodesRunning == co.numVirtualNodes {
		temp := big.NewInt(0)
		temp.SetBytes(co.localID)
		return co.localID, nil
	} else {
		return nil, fmt.Errorf("node ID not known yet")
	}
}

func (co *Chord) Leave(ctx context.Context) error {
	err := co.chordRing.Leave()
	if err != nil {
		return fmt.Errorf("leave chord error: %s", err)
	}
	return nil
}
