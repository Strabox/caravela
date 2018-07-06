package common

import (
	"github.com/strabox/caravela/node/common/guid"
)

/*
Represents a remote CARAVELA node.
*/
type RemoteNode struct {
	ip   string
	guid *guid.GUID
}

func NewRemoteNode(ipAddress string, guid guid.GUID) *RemoteNode {
	res := &RemoteNode{}
	res.ip = ipAddress
	res.guid = &guid
	return res
}

func (rm *RemoteNode) IP() string {
	return rm.ip
}

func (rm *RemoteNode) GUID() *guid.GUID {
	return rm.guid.Copy()
}
