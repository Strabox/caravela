package configuration

import (
	"github.com/strabox/caravela/node/common/resources"
	"time"
)

/*
Used to configure CARAVELA's node. Static parameters during system execution.
*/
type Configuration struct {
	hostIP           string
	dockerAPIVersion string

	// CARAVELA system configurations
	checkContainersInterval time.Duration
	supplyingInterval       time.Duration
	refreshesCheckInterval  time.Duration
	refreshingInterval      time.Duration
	maxRefreshesFailed      int
	maxRefreshesMissed      int
	refreshMissedTimeout    time.Duration

	// Overlay (chord) configurations
	overlayPort        int
	chordTimeout       time.Duration
	chordVirtualNodes  int
	chordNumSuccessors int
	chordHashSizeBits  int

	// Resources Partitions
	cpuPartitions []resources.ResourcePartition
	ramPartitions []resources.ResourcePartition

	// CARAVELA's REST API configurations
	apiPort    int
	apiTimeout time.Duration
}

func Default(hostIP string, dockerAPIVersion string) *Configuration {
	res := &Configuration{}
	res.hostIP = hostIP
	res.dockerAPIVersion = dockerAPIVersion

	res.checkContainersInterval = 30 * time.Second
	res.supplyingInterval = 45 * time.Second
	res.refreshesCheckInterval = 30 * time.Second
	res.refreshingInterval = 15 * time.Second
	res.maxRefreshesFailed = 2
	res.maxRefreshesMissed = 2
	res.refreshMissedTimeout = res.refreshingInterval + (5 * time.Second)

	res.overlayPort = 8000
	res.chordTimeout = 5 * time.Second
	res.chordVirtualNodes = 6
	res.chordNumSuccessors = 3
	res.chordHashSizeBits = 160

	res.cpuPartitions = []resources.ResourcePartition{{Value: 1, Percentage: 50}, {Value: 2, Percentage: 50}}
	res.ramPartitions = []resources.ResourcePartition{{Value: 256, Percentage: 50}, {Value: 512, Percentage: 50}}

	res.apiPort = 8001
	res.apiTimeout = 2 * time.Second
	return res
}

func (c *Configuration) HostIP() string {
	return c.hostIP
}

func (c *Configuration) DockerAPIVersion() string {
	return c.dockerAPIVersion
}

func (c *Configuration) CheckContainersInterval() time.Duration {
	return c.checkContainersInterval
}

func (c *Configuration) SupplyingInterval() time.Duration {
	return c.supplyingInterval
}

func (c *Configuration) RefreshesCheckInterval() time.Duration {
	return c.refreshesCheckInterval
}

func (c *Configuration) RefreshingInterval() time.Duration {
	return c.refreshingInterval
}

func (c *Configuration) MaxRefreshesMissed() int {
	return c.maxRefreshesMissed
}

func (c *Configuration) MaxRefreshesFailed() int {
	return c.maxRefreshesFailed
}

func (c *Configuration) RefreshMissedTimeout() time.Duration {
	return c.refreshMissedTimeout
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

func (c *Configuration) CpuPartitions() []resources.ResourcePartition {
	return c.cpuPartitions
}

func (c *Configuration) RamPartitions() []resources.ResourcePartition {
	return c.ramPartitions
}

func (c *Configuration) APIPort() int {
	return c.apiPort
}

func (c *Configuration) APITimeout() time.Duration {
	return c.apiTimeout
}
