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

func (r *RemoteNode) IP() string {
	return r.ip
}

func (r *RemoteNode) GUID() *guid.GUID {
	return r.guid.Copy()
}
