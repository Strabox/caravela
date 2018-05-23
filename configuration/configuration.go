package configuration

import (
	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"time"
)

const configurationFilePath = "configuration.toml"

/*
CARAVELA system's configurations
*/
type Configuration struct {
	Host     host
	Caravela caravela
	Overlay  overlay
}

/*
Configurations for the local host node
*/
type host struct {
	IP               string // Local node host's IP
	DockerAPIVersion string // API Version of the local node Docker's engine
}

/*
Configurations for the CARAVELA's node specific parameters
*/
type caravela struct {
	ApiPort                 int                 // Port of API REST endpoints
	ApiTimeout              duration            // Timeout for API REST requests
	MaxRefreshesFailed      int                 // Maximum amount of refreshes that a supplier failed to reply
	MaxRefreshesMissed      int                 // Maximum amount of refreshes a trader failed to send to the supplier
	CheckContainersInterval duration            // Interval of time to check the containers running in the node
	SupplyingInterval       duration            // Interval for supplier to check if it is necessary offer resources
	RefreshesCheckInterval  duration            // Interval to check if the refreshes to its offers are being done
	RefreshingInterval      duration            // Interval for trader to send refresh messages to suppliers
	RefreshMissedTimeout    duration            // Timeout for a refresh message
	CpuPowerPartition       []CpuPowerPartition // GUID partitions for CPU power
	CpuCoresPartitions      []CpuCoresPartition // GUID partitions for the amount of CPU cores
	RamPartitions           []RamPartition      // GUID partitions for the amount of ram
}

/*
Configurations for the node overlay
*/
type overlay struct {
	OverlayPort        int      // Port of the overlay endpoints
	ChordTimeout       duration // Timeout for the chord messages
	ChordVirtualNodes  int      // Number of chord virtual nodes per physical node
	ChordNumSuccessors int      // Number of chord successor nodes for a node
	ChordHashSizeBits  int      // Number of chord hash size (in bits)
}

/*
Produces the configuration structure with all the default values for the system to work.
*/
func defaultConfig(hostIP string, dockerAPIVersion string) *Configuration {
	config := &Configuration{}

	// Host
	config.Host.IP = hostIP
	config.Host.DockerAPIVersion = dockerAPIVersion

	// Caravela
	config.Caravela.ApiPort = 8000
	config.Caravela.ApiTimeout = duration{Duration: 2 * time.Second}
	config.Caravela.CheckContainersInterval = duration{Duration: 30 * time.Second}
	config.Caravela.SupplyingInterval = duration{Duration: 45 * time.Second}
	config.Caravela.RefreshesCheckInterval = duration{Duration: 30 * time.Second}
	config.Caravela.RefreshingInterval = duration{Duration: 15 * time.Second}
	config.Caravela.MaxRefreshesFailed = 2
	config.Caravela.MaxRefreshesMissed = 2
	config.Caravela.RefreshMissedTimeout = duration{Duration: config.Caravela.RefreshingInterval.Duration + (5 * time.Second)}
	config.Caravela.CpuCoresPartitions = []CpuCoresPartition{
		{Cores: 1, Percentage: 50}, {Cores: 2, Percentage: 50}}
	config.Caravela.RamPartitions = []RamPartition{
		{Ram: 256, Percentage: 50}, {Ram: 512, Percentage: 50}}

	// Overlay
	config.Overlay.OverlayPort = 8000
	config.Overlay.ChordTimeout = duration{Duration: 5 * time.Second}
	config.Overlay.ChordVirtualNodes = 3
	config.Overlay.ChordNumSuccessors = 3
	config.Overlay.ChordHashSizeBits = 160

	return config
}

/*
Produces configuration structure reading from the configuration file and filling the rest
with the default values
*/
func ReadConfigurations(hostIP string, dockerAPIVersion string) *Configuration {
	config := defaultConfig(hostIP, dockerAPIVersion)
	if _, err := toml.DecodeFile(configurationFilePath, config); err != nil {
		log.Errorf("Error reading configuration file: %s", err)
	}
	return config
}

func (c *Configuration) HostIP() string {
	return c.Host.IP
}

func (c *Configuration) DockerAPIVersion() string {
	return c.Host.DockerAPIVersion
}

func (c *Configuration) APIPort() int {
	return c.Caravela.ApiPort
}

func (c *Configuration) APITimeout() time.Duration {
	return c.Caravela.ApiTimeout.Duration
}

func (c *Configuration) CheckContainersInterval() time.Duration {
	return c.Caravela.CheckContainersInterval.Duration
}

func (c *Configuration) SupplyingInterval() time.Duration {
	return c.Caravela.SupplyingInterval.Duration
}

func (c *Configuration) RefreshesCheckInterval() time.Duration {
	return c.Caravela.RefreshesCheckInterval.Duration
}

func (c *Configuration) RefreshingInterval() time.Duration {
	return c.Caravela.RefreshingInterval.Duration
}

func (c *Configuration) MaxRefreshesMissed() int {
	return c.Caravela.MaxRefreshesMissed
}

func (c *Configuration) MaxRefreshesFailed() int {
	return c.Caravela.MaxRefreshesFailed
}

func (c *Configuration) RefreshMissedTimeout() time.Duration {
	return c.Caravela.RefreshMissedTimeout.Duration
}

func (c *Configuration) CpuPowerPartitions() []CpuPowerPartition {
	return c.Caravela.CpuPowerPartition
}

func (c *Configuration) CpuCoresPartitions() []CpuCoresPartition {
	return c.Caravela.CpuCoresPartitions
}

func (c *Configuration) RamPartitions() []RamPartition {
	return c.Caravela.RamPartitions
}

func (c *Configuration) OverlayPort() int {
	return c.Overlay.OverlayPort
}

func (c *Configuration) ChordTimeout() time.Duration {
	return c.Overlay.ChordTimeout.Duration
}

func (c *Configuration) ChordVirtualNodes() int {
	return c.Overlay.ChordVirtualNodes
}

func (c *Configuration) ChordNumSuccessors() int {
	return c.Overlay.ChordNumSuccessors
}

func (c *Configuration) ChordHashSizeBits() int {
	return c.Overlay.ChordHashSizeBits
}
