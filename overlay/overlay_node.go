package overlay

/*
Represents an overlay node.
*/
type Node struct {
	nodeIP string
	guid   []byte
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
