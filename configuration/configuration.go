package configuration

import (
	"fmt"
	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/util"
	"strings"
	"time"
)

// Minimum API version of docker engine supported
const minimumDockerEngineVersion = "1.35"

// Default port for the CARAVELA's API endpoints
const caravelaAPIPort = 8001

// Directory path to where search for the configuration file. (Directory of binary execution)
const configurationFilePath = ""

// Expected name of the configuration file.
const configurationFileName = "configuration.toml"

// CARAVELA system's configurations.
type Configuration struct {
	Host          host          `json:"-"` // Do not encode host configuration due to security concerns!!!
	Caravela      caravela      `json:"Caravela"`
	ImagesStorage imagesStorage `json:"ImagesStorage"`
	Overlay       overlay       `json:"Overlay"`
}

// Configurations for the local host node.
type host struct {
	IP               string `json:"IP"`               // Local node host's IP
	DockerAPIVersion string `json:"DockerAPIVersion"` // API Version of the local node Docker's engine
}

// Configurations for the CARAVELA's node specific parameters
type caravela struct {
	Simulation              bool                `json:"APIPort"`                 // If the CARAVELA node is simulated or not
	APIPort                 int                 `json:"APIPort"`                 // Port of API REST endpoints
	APITimeout              duration            `json:"APITimeout"`              // Timeout for API REST requests
	MaxRefreshesFailed      int                 `json:"MaxRefreshesFailed"`      // Maximum amount of refreshes that a supplier failed to reply
	MaxRefreshesMissed      int                 `json:"MaxRefreshesMissed"`      // Maximum amount of refreshes a trader failed to send to the supplier
	OffersStrategy          string              `json:"OffersStrategy"`          // Define what strategy is used to manage the offers
	CheckContainersInterval duration            `json:"CheckContainersInterval"` // Interval of time to check the containers running in the node
	SupplyingInterval       duration            `json:"SupplyingInterval"`       // Interval for supplier to check if it is necessary offer resources
	SpreadOffersInterval    duration            `json:"SpreadOffersInterval"`    // Interval for the trader to spread offer information into neighbors
	RefreshesCheckInterval  duration            `json:"RefreshesCheckInterval"`  // Interval to check if the refreshes to its offers are being done
	RefreshingInterval      duration            `json:"RefreshingInterval"`      // Interval for trader to send refresh messages to suppliers
	RefreshMissedTimeout    duration            `json:"RefreshMissedTimeout"`    // Timeout for a refresh message
	CPUPowerPartition       []CPUPowerPartition `json:"CPUPowerPartition"`       // GUID partitions for CPU power
	CPUCoresPartitions      []CPUCoresPartition `json:"CPUCoresPartitions"`      // GUID partitions for the amount of CPU cores
	RAMPartitions           []RAMPartition      `json:"RAMPartitions"`           // GUID partitions for the amount of RAM
	ResourcesOvercommit     int                 `json:"ResourcesOvercommit"`     // Percentage of overcommit to apply to available resources
}

// Configuration for the CARAVELA's container image storage
type imagesStorage struct {
	Backend string `json:"Backend"` // Type of storage of images used to share them
}

// Configurations for the node overlay
type overlay struct {
	OverlayPort        int      `json:"OverlayPort"`        // Port of the overlay endpoints
	ChordTimeout       duration `json:"ChordTimeout"`       // Timeout for the chord messages
	ChordVirtualNodes  int      `json:"ChordVirtualNodes"`  // Number of chord virtual nodes per physical node
	ChordNumSuccessors int      `json:"ChordNumSuccessors"` // Number of chord successor nodes for a node
	ChordHashSizeBits  int      `json:"ChordHashSizeBits"`  // Number of chord hash size (in bits)
}

// Produces the configuration structure with all the default values for the system to work.
func Default(hostIP string) *Configuration {
	refreshingInterval := duration{Duration: 15 * time.Second}

	return &Configuration{
		Host: host{
			IP:               hostIP,
			DockerAPIVersion: minimumDockerEngineVersion,
		},
		Caravela: caravela{
			Simulation:              false,
			APIPort:                 caravelaAPIPort,
			OffersStrategy:          "chordDefault",
			APITimeout:              duration{Duration: 5 * time.Second},
			CheckContainersInterval: duration{Duration: 30 * time.Second},
			SupplyingInterval:       duration{Duration: 45 * time.Second},
			RefreshesCheckInterval:  duration{Duration: 30 * time.Second},
			RefreshingInterval:      refreshingInterval,
			MaxRefreshesFailed:      2,
			MaxRefreshesMissed:      2,
			RefreshMissedTimeout:    duration{Duration: refreshingInterval.Duration + (5 * time.Second)},
			CPUPowerPartition: []CPUPowerPartition{
				{Class: 1, ResourcesPartition: ResourcesPartition{Percentage: 50}},
				{Class: 2, ResourcesPartition: ResourcesPartition{Percentage: 50}}},
			CPUCoresPartitions: []CPUCoresPartition{
				{Cores: 1, ResourcesPartition: ResourcesPartition{Percentage: 50}},
				{Cores: 2, ResourcesPartition: ResourcesPartition{Percentage: 50}}},
			RAMPartitions: []RAMPartition{
				{RAM: 256, ResourcesPartition: ResourcesPartition{Percentage: 50}},
				{RAM: 512, ResourcesPartition: ResourcesPartition{Percentage: 50}}},
			ResourcesOvercommit: 50,
		},
		ImagesStorage: imagesStorage{
			Backend: ImagesStorageDockerHub,
		},
		Overlay: overlay{
			OverlayPort:        8000,
			ChordTimeout:       duration{Duration: 5 * time.Second},
			ChordVirtualNodes:  3,
			ChordNumSuccessors: 3,
			ChordHashSizeBits:  160,
		},
	}
}

// Produces configuration structure reading from the configuration file and filling the rest
// with the default values
func ReadFromFile(hostIP string) (*Configuration, error) {
	config := Default(hostIP)
	configFullFileName := configurationFilePath + configurationFileName

	if _, err := toml.DecodeFile(configFullFileName, config); err != nil && !strings.Contains(err.Error(), "cannot find the file") {
		return nil, err
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Produces configuration structured based on a given structure that.
// Used to pass the system configurations between nodes, usually during the joining process.
func ObtainExternal(hostIP string, config *Configuration) (*Configuration, error) {
	config.Host.IP = hostIP

	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Briefly validate the configuration to avoid/short-circuit many runtime errors due to
// typos or completely non sense configurations, like negative ports.
func (c *Configuration) validate() error {
	/* TODO: Undo this when simulator generates valid unique IPs
	if net.ParseIP(c.HostIP()) == nil {
		return fmt.Errorf("invalid host ip address: %s", c.HostIP())
	}
	*/
	if !util.IsValidPort(c.APIPort()) {
		return fmt.Errorf("invalid api port: %d", c.APIPort())
	}
	if c.MaxRefreshesFailed() < 0 {
		return fmt.Errorf("maximum number of failed refreshes must be a positive integer")
	}
	if c.MaxRefreshesMissed() < 0 {
		return fmt.Errorf("maximum number of missed refreshes must be a positive integer")
	}
	if c.ResourcesOvercommit() <= 0 {
		return fmt.Errorf("node's resources overcommit ratio must be a positive integer")
	}

	percentageAcc := 0
	for _, value := range c.CPUPowerPartitions() {
		percentageAcc += value.Percentage
		if value.Class < 0 {
			return fmt.Errorf("partitions CPU power class must be a positive integer")
		}
	}
	if percentageAcc != 100 {
		return fmt.Errorf("the sum of CPU power partitions size must equal to 100")
	}

	percentageAcc = 0
	for _, value := range c.CPUCoresPartitions() {
		percentageAcc += value.Percentage
		if value.Cores <= 0 {
			return fmt.Errorf("partitions CPU cores must be a positive integer")
		}
	}
	if percentageAcc != 100 {
		return fmt.Errorf("the sum of CPU cores partitions size must equal to 100")
	}

	percentageAcc = 0
	for _, value := range c.RAMPartitions() {
		percentageAcc += value.Percentage
		if value.RAM <= 0 {
			return fmt.Errorf("partitions RAM amount must be a positive integer")
		}
	}
	if percentageAcc != 100 {
		return fmt.Errorf("the sum of RAM partitions size must equal to 100")
	}

	configuredBackend := strings.ToLower(c.ImagesStorage.Backend)
	if configuredBackend != strings.ToLower(ImagesStorageDockerHub) &&
		configuredBackend != strings.ToLower(ImagesStorageIPFS) {
		return fmt.Errorf("invalid storage backend: %s", configuredBackend)
	}

	if !util.IsValidPort(c.OverlayPort()) {
		return fmt.Errorf("invalid overlay port: %d", c.OverlayPort())
	}
	if c.ChordVirtualNodes() <= 0 {
		return fmt.Errorf("chord's number of virtual nodes must be a positive integer")
	}
	if c.ChordNumSuccessors() <= 0 {
		return fmt.Errorf("chord's number of successor nodes must be a positive integer")
	}
	if c.ChordHashSizeBits() < 56 {
		return fmt.Errorf("chord's hash size bits nodes must be a positive integer greater or equal to 56")
	}

	return nil
}

// Print/log the current configurations in order to debug the programs behavior.
func (c *Configuration) Print() {
	log.Printf("##################################################################")
	log.Printf("#                       CONFIGURATIONS                           #")
	log.Printf("##################################################################")

	log.Printf("$$$$$$$$$$$$$$$$$$$$$$$$$$$ HOST $$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
	log.Printf("IP Address:                  %s", c.HostIP())
	log.Printf("Docker Engine Version:       %s", c.DockerAPIVersion())

	log.Printf("$$$$$$$$$$$$$$$$$$$$$$$$$$ CARAVELA $$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
	log.Printf("Simulation:                  %t", c.Simulation())
	log.Printf("Port:                        %d", c.APIPort())
	log.Printf("Messages Timeout:            %s", c.APITimeout().String())
	log.Printf("OffersStrategy:              %s", c.OffersStrategy())
	log.Printf("Check Containers Interval:   %s", c.CheckContainersInterval().String())
	log.Printf("Supply Resources Interval:   %s", c.SupplyingInterval().String())
	log.Printf("Refreshes Check Interval:    %s", c.RefreshesCheckInterval().String())
	log.Printf("Refreshes Interval:          %s", c.RefreshingInterval().String())
	log.Printf("Refresh missed timeout:      %s", c.RefreshMissedTimeout().String())
	log.Printf("Max num of refreshes failed: %d", c.MaxRefreshesFailed())
	log.Printf("Max num of refreshes missed: %d", c.MaxRefreshesMissed())
	log.Printf("Resources overcommit:        %d", c.ResourcesOvercommit())

	partitions := ""
	for i, value := range c.CPUPowerPartitions() {
		if i > 0 {
			partitions += ", "
		}
		partitions += fmt.Sprintf("<%d,%d>", value.Class, value.Percentage)
	}
	log.Printf("CPU Power Partitions:        %s", partitions)

	partitions = ""
	for i, value := range c.CPUCoresPartitions() {
		if i > 0 {
			partitions += ", "
		}
		partitions += fmt.Sprintf("<%d,%d>", value.Cores, value.Percentage)
	}
	log.Printf("CPU Cores Partitions:        %s", partitions)

	partitions = ""
	for i, value := range c.RAMPartitions() {
		if i > 0 {
			partitions += ", "
		}
		partitions += fmt.Sprintf("<%d,%d>", value.RAM, value.Percentage)
	}
	log.Printf("RAM Partitions:              %s", partitions)

	log.Printf("$$$$$$$$$$$$$$$$$$$$$$$ IMAGES STORAGE $$$$$$$$$$$$$$$$$$$$$$$$$$$")
	log.Printf("Backend:                     %s", c.ImagesStorageBackend())

	log.Printf("$$$$$$$$$$$$$$$$$$$$$$$$$$ OVERLAY $$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
	log.Printf("Port:                        %d", c.OverlayPort())
	log.Printf("Messages Timeout:            %s", c.ChordTimeout().String())
	log.Printf("Number of Virtual Nodes:     %d", c.ChordVirtualNodes())
	log.Printf("Number of Successors:        %d", c.ChordNumSuccessors())
	log.Printf("Hash Size (bits):            %d", c.ChordHashSizeBits())

	log.Printf("##################################################################")
}

func (c *Configuration) HostIP() string {
	return c.Host.IP
}

func (c *Configuration) DockerAPIVersion() string {
	return c.Host.DockerAPIVersion
}

func (c *Configuration) ImagesStorageBackend() string {
	return c.ImagesStorage.Backend
}

func (c *Configuration) Simulation() bool {
	return c.Caravela.Simulation
}

func (c *Configuration) APIPort() int {
	return c.Caravela.APIPort
}

func (c *Configuration) APITimeout() time.Duration {
	return c.Caravela.APITimeout.Duration
}

func (c *Configuration) CheckContainersInterval() time.Duration {
	return c.Caravela.CheckContainersInterval.Duration
}

func (c *Configuration) SpreadOffersInterval() time.Duration {
	return c.Caravela.SpreadOffersInterval.Duration
}

func (c *Configuration) OffersStrategy() string {
	return c.Caravela.OffersStrategy
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

func (c *Configuration) CPUPowerPartitions() []CPUPowerPartition {
	return c.Caravela.CPUPowerPartition
}

func (c *Configuration) CPUCoresPartitions() []CPUCoresPartition {
	return c.Caravela.CPUCoresPartitions
}

func (c *Configuration) RAMPartitions() []RAMPartition {
	return c.Caravela.RAMPartitions
}

func (c *Configuration) ResourcesOvercommit() int {
	return c.Caravela.ResourcesOvercommit
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
