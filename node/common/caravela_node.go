package common

import (
	"github.com/strabox/caravela/node/common/guid"
)

// RemoteNode represents a remote CARAVELA node.
type RemoteNode struct {
	ip   string
	guid *guid.GUID
}

func NewRemoteNode(IPAddress string, guid guid.GUID) *RemoteNode {
	return &RemoteNode{
		ip:   IPAddress,
		guid: &guid,
	}
}

func (rm *RemoteNode) IP() string {
	return rm.ip
}

func (rm *RemoteNode) GUID() *guid.GUID {
	return rm.guid.Copy()
}
