package overlay

type RemoteNode struct {
	nodeIP string
	guid   []byte
}

func NewRemoteNode(nodeIP string, guid []byte) *RemoteNode {
	res := &RemoteNode{}
	res.nodeIP = nodeIP
	res.guid = guid
	return res
}

func (rn *RemoteNode) IP() string {
	return rn.nodeIP
}

func (rn *RemoteNode) Guid() []byte {
	return rn.guid
}
