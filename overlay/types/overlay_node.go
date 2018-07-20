package types

// Represents a generic overlay node.
type OverlayNode struct {
	// IP address
	nodeIP string
	// Port where it runs the overlay daemon
	port int
	// Node identifier
	guid []byte
}

// NewOverlayNode creates a new overlay node.
func NewOverlayNode(nodeIP string, port int, guid []byte) *OverlayNode {
	return &OverlayNode{
		nodeIP: nodeIP,
		port:   port,
		guid:   guid,
	}
}

// IP of the node
func (rn *OverlayNode) IP() string {
	return rn.nodeIP
}

// Port of the overlay daemon
func (rn *OverlayNode) Port() int {
	return rn.port
}

// Node identifier
func (rn *OverlayNode) GUID() []byte {
	return rn.guid
}
