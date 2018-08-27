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
func (o *OverlayNode) IP() string {
	return o.nodeIP
}

// Port of the overlay daemon
func (o *OverlayNode) Port() int {
	return o.port
}

// Node identifier
func (o *OverlayNode) GUID() []byte {
	return o.guid
}
