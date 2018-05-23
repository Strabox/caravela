package overlay

/*
Represents a generic and abstract overlay node.
*/
type Node struct {
	// IP address
	nodeIP string
	// System's node identifier
	guid []byte
}

func NewNode(nodeIP string, guid []byte) *Node {
	res := &Node{}
	res.nodeIP = nodeIP
	res.guid = guid
	return res
}

func (rn *Node) IP() string {
	return rn.nodeIP
}

func (rn *Node) GUID() []byte {
	return rn.guid
}
