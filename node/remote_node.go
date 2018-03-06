package node

import ()

type RemoteNode struct {
	nodeIP string
	guid   *Guid
}

func NewRemoteNode(nodeIP string, guid *Guid) *RemoteNode {
	res := &RemoteNode{}
	res.nodeIP = nodeIP
	res.guid = guid
	return res
}
