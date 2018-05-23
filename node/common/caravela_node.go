package common

import (
	"github.com/strabox/caravela/node/common/guid"
)

/*
Represents a remote CARAVELA node.
*/
type RemoteNode struct {
	ipAddress string
	guid      *guid.GUID
}

func NewRemoteNode(ipAddress string, guid guid.GUID) *RemoteNode {
	res := &RemoteNode{}
	res.ipAddress = ipAddress
	res.guid = &guid
	return res
}

func (rm *RemoteNode) IPAddress() string {
	return rm.ipAddress
}

func (rm *RemoteNode) GUID() *guid.GUID {
	return rm.guid.Copy()
}
