package configuration

import (
	"time"
	"github.com/strabox/caravela/node/resources"
)

/*
Used to configurate  CARAVELA's node.
Static parameters.
*/
type Configuration struct {
	HostIP				string
	SupplyingInterval	time.Duration

	// Overlay (chord) configurations
	OverlayPort 		int
	ChordTimeout 		time.Duration
	ChordVirtualNodes 	int
	ChordNumSuccessors	int
	ChordHashSizeBits	int
	
	// Resources Partitions
	CpuPartitions		[]resources.ResourcePerc
	RamPartitions		[]resources.ResourcePerc
	
	// CARAVELA's REST API configurations
	APIPort 			int
	APITimeout 			time.Duration
}


func DefaultConfiguration(hostIP string) *Configuration {
	res := &Configuration{}
	res.HostIP = hostIP
	res.SupplyingInterval = 10 * time.Second
		
	res.OverlayPort = 8000
	res.ChordTimeout = 2 * time.Second
	res.ChordVirtualNodes = 6
	res.ChordNumSuccessors = 3
	res.ChordHashSizeBits = 160
	
	res.CpuPartitions = []resources.ResourcePerc{resources.ResourcePerc{1,50}, resources.ResourcePerc{2,50}}
	res.RamPartitions = []resources.ResourcePerc{resources.ResourcePerc{256,50}, resources.ResourcePerc{512,50}}
	
	res.APIPort = 8001
	res.APITimeout = 2 * time.Second
	return res
}