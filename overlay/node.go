package overlay

/*
Represents a generic overlay node.
*/
type Node struct {
	// IP address
	nodeIP string
	// Port
	port int
	// System's node identifier
	guid []byte
}

func NewNode(nodeIP string, port int, guid []byte) *Node {
	res := &Node{}
	res.nodeIP = nodeIP
	res.port = port
	res.guid = guid
	return res
}

func (rn *Node) IP() string {
	return rn.nodeIP
}

func (rn *Node) Port() int {
	return rn.port
}

func (rn *Node) GUID() []byte {
	return rn.guid
}
