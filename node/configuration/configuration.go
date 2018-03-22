package configuration

import (
	"github.com/strabox/caravela/node/resources"
	"time"
)

/*
Used to configure  CARAVELA's node. Static parameters during system execution.
*/
type Configuration struct {
	hostIP             string
	supplyingInterval  time.Duration
	refreshingInterval time.Duration

	// Overlay (chord) configurations
	overlayPort        int
	chordTimeout       time.Duration
	chordVirtualNodes  int
	chordNumSuccessors int
	chordHashSizeBits  int

	// Resources Partitions
	cpuPartitions []resources.ResourcePerc
	ramPartitions []resources.ResourcePerc

	// CARAVELA's REST API configurations
	apiPort    int
	apiTimeout time.Duration
}

func DefaultConfiguration(hostIP string) *Configuration {
	res := &Configuration{}
	res.hostIP = hostIP
	res.supplyingInterval = 10 * time.Second
	res.refreshingInterval = 15 * time.Second

	res.overlayPort = 8000
	res.chordTimeout = 2 * time.Second
	res.chordVirtualNodes = 6
	res.chordNumSuccessors = 3
	res.chordHashSizeBits = 160

	res.cpuPartitions = []resources.ResourcePerc{{1, 50}, {2, 50}}
	res.ramPartitions = []resources.ResourcePerc{{256, 50}, {512, 50}}

	res.apiPort = 8001
	res.apiTimeout = 2 * time.Second
	return res
}

func (c *Configuration) HostIP() string {
	return c.hostIP
}

func (c *Configuration) SupplyingInterval() time.Duration {
	return c.supplyingInterval
}

func (c *Configuration) RefreshingInterval() time.Duration {
	return c.refreshingInterval
}

func (c *Configuration) OverlayPort() int {
	return c.overlayPort
}

func (c *Configuration) ChordTimeout() time.Duration {
	return c.chordTimeout
}

func (c *Configuration) ChordVirtualNodes() int {
	return c.chordVirtualNodes
}

func (c *Configuration) ChordNumSuccessors() int {
	return c.chordNumSuccessors
}

func (c *Configuration) ChordHashSizeBits() int {
	return c.chordHashSizeBits
}

func (c *Configuration) CpuPartitions() []resources.ResourcePerc {
	return c.cpuPartitions
}

func (c *Configuration) RamPartitions() []resources.ResourcePerc {
	return c.ramPartitions
}

func (c *Configuration) APIPort() int {
	return c.apiPort
}

func (c *Configuration) APITimeout() time.Duration {
	return c.apiTimeout
}
