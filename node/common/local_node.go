package common

import "github.com/strabox/caravela/node/discovery/offering/partitions"

type Node interface {
	GetSystemPartitionsState() *partitions.SystemResourcePartitions
	GUID() string
}
