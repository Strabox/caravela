package common

import (
	"github.com/strabox/caravela/node/common/guid"
)

/*
Represents a remote CARAVELA node.
*/
type RemoteNode struct {
	ipAddress string
	guid      *guid.Guid
}

func NewRemoteNode(ipAddress string, guid guid.Guid) *RemoteNode {
	res := &RemoteNode{}
	res.ipAddress = ipAddress
	res.guid = &guid
	return res
}

func (rm *RemoteNode) IPAddress() string {
	return rm.ipAddress
}

func (rm *RemoteNode) GUID() *guid.Guid {
	return rm.guid.Copy()
}
