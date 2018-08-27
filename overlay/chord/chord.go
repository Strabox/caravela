package chord

import (
	"context"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/overlay/types"
	"github.com/strabox/caravela/util"
	"github.com/strabox/go-chord"
	"hash"
	"math/big"
	"strconv"
	"sync"
	"time"
)

// Chord represents a Chord overlay local (for each node) structure.
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

// new create a new chord overlay structure.
func New(config *configuration.Configuration) (types.Overlay, error) {
	return &Chord{
		appNode:             nil,
		predecessors:        sync.Map{},
		virtualNodesRunning: 0,
		localID:             nil,

		hashSizeBytes:   config.ChordHashSizeBits() / 8,
		hostIP:          config.HostIP(),
		hostPort:        config.OverlayPort(),
		numVirtualNodes: config.ChordVirtualNodes(),
		numSuccessors:   config.ChordNumSuccessors(),
		timeout:         config.ChordTimeout(),
		chordRing:       nil,
	}, nil
}

// initialize the chord overlay and its respective inner structures.
func (c *Chord) initialize(appNode types.OverlayMembership) (*chord.Config, chord.Transport, error) {
	c.appNode = appNode
	hostname := c.hostIP + ":" + strconv.Itoa(c.hostPort)
	chordListener := &Listener{chordOverlay: c}

	config := chord.DefaultConfig(hostname)
	config.Delegate = chordListener
	config.NumVnodes = c.numVirtualNodes
	config.NumSuccessors = c.numSuccessors
	config.HashFunc = func() hash.Hash { return NewResourcesHash(c.hashSizeBytes, hostname) }

	// Initialize the TCP transport stack used in the chord implementation
	transport, err := chord.InitTCPTransport(fmt.Sprintf(":%d", c.hostPort), c.timeout)
	if err != nil {
		return nil, nil, err
	}

	return config, transport, nil
}

// Called when a new virtual node of this physical node has joined the chord ring.
func (c *Chord) newLocalVirtualNode(localVirtualNodeID []byte, predecessorNode *types.OverlayNode) {
	if c.virtualNodesRunning == c.numVirtualNodes {
		return
	}

	newLocalVirtualNodeID := big.NewInt(0)
	newLocalVirtualNodeID.SetBytes(localVirtualNodeID)
	if c.localID == nil {
		c.localID = localVirtualNodeID
	} else {
		localID := big.NewInt(0)
		localID.SetBytes(c.localID)
		if newLocalVirtualNodeID.Cmp(localID) >= 0 {
			c.localID = localVirtualNodeID
		}
	}
	c.virtualNodesRunning++
	c.predecessors.Store(newLocalVirtualNodeID.String(), predecessorNode)
	c.appNode.AddTrader(localVirtualNodeID) // Alert the node for the new virtual node/trader
}

// Called when the predecessor of a virtual node of the physical node changes.
// e.g. Due to a crash in the previous predecessor or because he left the chord ring.
func (c *Chord) predecessorNodeChanged(localVirtualNodeID []byte, predecessorNode *types.OverlayNode) {
	vNodeID := big.NewInt(0)
	vNodeID.SetBytes(localVirtualNodeID)
	c.predecessors.Store(vNodeID.String(), predecessorNode)
}

/* ============================ Overlay Interface ============================ */

func (c *Chord) Create(ctx context.Context, appNode types.OverlayMembership) error {
	config, transport, err := c.initialize(appNode)
	if err != nil {
		return err
	}

	ring, err := chord.Create(config, transport)
	if err != nil {
		return fmt.Errorf("create chord error: %s", err)
	}

	c.chordRing = ring
	return nil
}

func (c *Chord) Join(ctx context.Context, overlayNodeIP string, overlayNodePort int, appNode types.OverlayMembership) error {
	config, transport, err := c.initialize(appNode)
	if err != nil {
		return err
	}

	var joinHostname = overlayNodeIP + ":" + strconv.Itoa(overlayNodePort)
	ring, err := chord.Join(config, transport, joinHostname)
	if err != nil {
		return fmt.Errorf("join chord error: %s", err)
	}
	c.chordRing = ring
	return nil
}

func (c *Chord) Lookup(ctx context.Context, key []byte) ([]*types.OverlayNode, error) {
	virtualNodes, err := c.chordRing.Lookup(c.numSuccessors, key)
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

func (c *Chord) Neighbors(ctx context.Context, nodeID []byte) ([]*types.OverlayNode, error) {
	id := big.NewInt(0)
	id.SetBytes(nodeID)

	res := make([]*types.OverlayNode, 0)
	nodes, err := c.Lookup(ctx, nodeID) // TODO: Optional, avoid this lookup.
	if err != nil {
		return make([]*types.OverlayNode, 0), err
	}

	if len(nodes) > 1 {
		hmm := big.NewInt(0)
		hmm.SetBytes(nodes[1].GUID())
		res = append(res, nodes[1]) // The successor of the given node
	}
	predecessorNode, exist := c.predecessors.Load(id.String())
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

func (c *Chord) NodeID(ctx context.Context) ([]byte, error) {
	if c.localID != nil && c.virtualNodesRunning == c.numVirtualNodes {
		temp := big.NewInt(0)
		temp.SetBytes(c.localID)
		return c.localID, nil
	} else {
		return nil, fmt.Errorf("node ID not known yet")
	}
}

func (c *Chord) Leave(ctx context.Context) error {
	err := c.chordRing.Leave()
	if err != nil {
		return fmt.Errorf("leave chord error: %s", err)
	}
	return nil
}
