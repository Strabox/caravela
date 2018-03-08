package overlay

import (
	"github.com/strabox/caravela/node/guid"
)

type RemoteNode struct {
	nodeIP string
	guid   *guid.Guid
}

func NewRemoteNode(nodeIP string, guid *guid.Guid) *RemoteNode {
	res := &RemoteNode{}
	res.nodeIP = nodeIP
	res.guid = guid
	return res
}

func (rn *RemoteNode) IP() string {
	return rn.nodeIP
}

func (rn *RemoteNode) Guid() *guid.Guid {
	return rn.guid.Copy()
}
