package overlay

// Represents a generic overlay node.
type Node struct {
	// IP address
	nodeIP string
	// Port where it runs the overlay daemon
	port int
	// Node identifier
	guid []byte
}

// NewNode creates a new overlay node.
func NewNode(nodeIP string, port int, guid []byte) *Node {
	return &Node{
		nodeIP: nodeIP,
		port:   port,
		guid:   guid,
	}
}

// IP of the node
func (rn *Node) IP() string {
	return rn.nodeIP
}

// Port of the overlay daemon
func (rn *Node) Port() int {
	return rn.port
}

// Node identifier
func (rn *Node) GUID() []byte {
	return rn.guid
}
