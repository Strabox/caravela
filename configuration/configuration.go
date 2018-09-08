package configuration

import (
	"fmt"
	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/util"
	"net"
	"strings"
	"time"
)

// Minimum API version of docker engine supported
const minimumDockerEngineVersion = "1.35"

// Default port for the CARAVELA's API endpoints
const defaultCaravelaAPIPort = 8001

// Directory path to where search for the configuration file. (Directory of binary execution)
const DefaultFilePath = "configuration.toml"

// Configuration holds all the configurations parameters for the CARAVELA.
type Configuration struct {
	Host          host                 `json:"Host"`
	Caravela      caravela             `json:"Caravela"`
	ImagesStorage imagesStorageBackend `json:"ImagesStorage"`
	Overlay       overlay              `json:"Overlay"`
}

// ##################################################################################################
// #                                          Host                                                  #
// ##################################################################################################

// Configurations for the local host node.
type host struct {
	IP               string `json:"-"`                // Do not encode host IP due to security concerns!!!
	DockerAPIVersion string `json:"DockerAPIVersion"` // API Version of the local node Docker's engine
}

// ##################################################################################################
// #                                         Caravela                                               #
// ##################################################################################################

// Configurations for the CARAVELA's node specific parameters
type caravela struct {
	Simulation       bool                `json:"Simulation"`       // If the CARAVELA node is under simulation or not.
	DiscoveryBackend discoveryBackend    `json:"DiscoveryBackend"` // Define what strategy is used to manage the offers
	APIPort          int                 `json:"APIPort"`          // Port of API REST endpoints
	APITimeout       duration            `json:"APITimeout"`       // Timeout for API REST requests
	CPUSlices        int                 `json:"CPUSlices"`        // Number of equal slices for a CPU/Core e.g. 2
	CPUOvercommit    int                 `json:"CPUOvercommit"`    // CPU overcommit percentage e.g. 140%
	MemoryOvercommit int                 `json:"MemoryOvercommit"` // Memory overcommit percentage e.g. 120%
	Resources        ResourcesPartitions `json:"FreeResources"`    // FreeResources partitions
	SchedulingPolicy string              `json:"SchedulingPolicy"` // Scheduling policies used when several nodes are available.
}

type discoveryBackend struct {
	Backend              string                   `json:"StorageBackend"`       // Selected discovery backend.
	OfferingChordBackend offeringChordDiscBackend `json:"OfferingChordBackend"` // SmartChord discovery backend configs.
	RandomChordBackend   randomChordDiscBackend   `json:"RandomChordBackend"`   // RandomChord discovery backend configs.
}

type offeringChordDiscBackend struct {
	SupplyingInterval      duration `json:"SupplyingInterval"`      // Interval for supplier to check if it is necessary offer resources
	SpreadOffersInterval   duration `json:"SpreadOffersInterval"`   // Interval for the trader to spread offer information into neighbors
	RefreshesCheckInterval duration `json:"RefreshesCheckInterval"` // Interval to check if the refreshes to its offers are being done
	RefreshingInterval     duration `json:"RefreshingInterval"`     // Interval for trader to send refresh messages to suppliers
	RefreshMissedTimeout   duration `json:"RefreshMissedTimeout"`   // Timeout for a refresh message
	MaxRefreshesFailed     int      `json:"MaxRefreshesFailed"`     // Maximum amount of refreshes that a supplier failed to reply
	MaxRefreshesMissed     int      `json:"MaxRefreshesMissed"`     // Maximum amount of refreshes a trader failed to send to the supplier
	// Debug performance flags
	SpreadOffers             bool `json:"SpreadOffers"`             // Used to tell if the spread offers mechanism is used or not.
	SpreadPartitionsState    bool `json:"SpreadPartitionsState"`    // Used to tell if the spread partitions state is used or not.
	GUIDEstimatedNetworkSize int  `json:"GUIDEstimatedNetworkSize"` // Estimated network size to tune the GUID scale.
	GUIDScaleFactor          int  `json:"GUIDScaleFactor"`          // GUID scale's factor.
}

type randomChordDiscBackend struct {
	RandBackendMaxRetries int `json:"RandBackendMaxRetries"` // Maximum retries when discovering resources.
}

// ##################################################################################################
// #                                    Images Storage                                              #
// ##################################################################################################

// Configuration for the CARAVELA's container image storage
type imagesStorageBackend struct {
	StorageBackend string `json:"StorageBackend"` // Type of storage of images used to share them
}

// ##################################################################################################
// #                                        Overlay                                                 #
// ##################################################################################################

// Configurations for the node overlay
type overlay struct {
	Overlay     string `json:"Overlay"`     // Overlay configured
	OverlayPort int    `json:"OverlayPort"` // Port of the overlay endpoints
	Chord       chord  `json:"Chord"`       // overlay configurations
}

type chord struct {
	Timeout       duration `json:"Timeout"`       // Timeout for the chord messages
	VirtualNodes  int      `json:"VirtualNodes"`  // Number of chord virtual nodes per physical node
	NumSuccessors int      `json:"NumSuccessors"` // Number of chord successor nodes for a node
	HashSizeBits  int      `json:"HashSizeBits"`  // Number of chord hash size (in bits)
}

// Default returns the configuration structure with all the default values for the system to work.
func Default(hostIP string) *Configuration {
	refreshingInterval := duration{Duration: 15 * time.Second}

	return &Configuration{
		Host: host{
			IP:               hostIP,
			DockerAPIVersion: minimumDockerEngineVersion,
		},
		Caravela: caravela{
			Simulation:       false,
			APIPort:          defaultCaravelaAPIPort,
			APITimeout:       duration{Duration: 5 * time.Second},
			CPUSlices:        1,
			CPUOvercommit:    100,
			MemoryOvercommit: 100,
			SchedulingPolicy: "binpack",
			DiscoveryBackend: discoveryBackend{
				Backend: "chord-single-offer",
				OfferingChordBackend: offeringChordDiscBackend{
					SupplyingInterval:      duration{Duration: 45 * time.Second},
					SpreadOffersInterval:   duration{Duration: 40 * time.Second},
					RefreshesCheckInterval: duration{Duration: 30 * time.Second},
					RefreshingInterval:     refreshingInterval,
					MaxRefreshesFailed:     2,
					MaxRefreshesMissed:     2,
					RefreshMissedTimeout:   duration{Duration: refreshingInterval.Duration + (5 * time.Second)},
					// Debug performance flags
					SpreadOffers:             true,
					SpreadPartitionsState:    true,
					GUIDEstimatedNetworkSize: 50000,
					GUIDScaleFactor:          5,
				},
				RandomChordBackend: randomChordDiscBackend{
					RandBackendMaxRetries: 6,
				},
			},
			Resources: ResourcesPartitions{
				CPUClasses: []CPUClassPartition{
					{
						ResourcesPartition: ResourcesPartition{Value: 0, Percentage: 100},
						CPUCores: []CPUCoresPartition{
							{
								ResourcesPartition: ResourcesPartition{Value: 1, Percentage: 50},
								Memory: []MemoryPartition{
									{ResourcesPartition: ResourcesPartition{Value: 256, Percentage: 25}},
									{ResourcesPartition: ResourcesPartition{Value: 512, Percentage: 50}},
									{ResourcesPartition: ResourcesPartition{Value: 1024, Percentage: 25}},
								},
							},
							{
								ResourcesPartition: ResourcesPartition{Value: 2, Percentage: 50},
								Memory: []MemoryPartition{
									{ResourcesPartition: ResourcesPartition{Value: 512, Percentage: 25}},
									{ResourcesPartition: ResourcesPartition{Value: 1024, Percentage: 50}},
									{ResourcesPartition: ResourcesPartition{Value: 2048, Percentage: 25}},
								},
							},
						},
					},
				},
			},
		},
		ImagesStorage: imagesStorageBackend{
			StorageBackend: "DockerHub",
		},
		Overlay: overlay{
			Overlay:     "chord",
			OverlayPort: 8000,
			Chord: chord{
				Timeout:       duration{Duration: 5 * time.Second},
				VirtualNodes:  3,
				NumSuccessors: 3,
				HashSizeBits:  160,
			},
		},
	}
}

// ReadFromFile returns a configuration structure reading from the configuration file and filling the rest
// with the default values
func ReadFromFile(hostIP, configFilePath string) (*Configuration, error) {
	config := Default(hostIP)

	if _, err := toml.DecodeFile(configFilePath, config); err != nil && strings.Contains(err.Error(), "cannot find the file") {
		return config, fmt.Errorf("cannot find the file %s", configFilePath)
	} else if err != nil {
		return nil, err
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// ObtainExternal returns configuration structured based on a given structure that.
// Used to pass the system configurations between nodes, usually during the joining process.
func ObtainExternal(hostIP string, config *Configuration) (*Configuration, error) {
	res := *config
	res.Host.IP = hostIP

	if err := res.validate(); err != nil {
		return nil, err
	}

	return &res, nil
}

// validate validates the configuration to avoid/short-circuit many runtime errors due to
// typos or completely non sense configurations, like negative ports.
func (c *Configuration) validate() error {
	// ==================================== Host ==============================================

	if net.ParseIP(c.HostIP()) == nil {
		return fmt.Errorf("invalid host ip address: %s", c.HostIP())
	}

	// =================================== Caravela ===========================================

	if !util.IsValidPort(c.APIPort()) {
		return fmt.Errorf("invalid backend port: %d", c.APIPort())
	}

	if c.CPUSlices() <= 0 {
		return fmt.Errorf("CPUSlices: %d, it must be >= 1", c.CPUSlices())
	}

	if c.CPUOvercommit() < 100 {
		return fmt.Errorf("CPUOvercommit: %d, CPU overcommit percentage must be >= 100", c.CPUOvercommit())
	}

	if c.MemoryOvercommit() < 100 {
		return fmt.Errorf("MemoryOvercommit: %d, Memory overcommit percentage must be >= 100", c.MemoryOvercommit())
	}

	powerPercentageAcc := 0
	for _, powerPart := range c.Caravela.Resources.CPUClasses {
		powerPercentageAcc += powerPart.Percentage
		currCoresPercentageAcc := 0
		for _, corePart := range powerPart.CPUCores {
			currCoresPercentageAcc += corePart.Percentage
			currMemoryPercentageAcc := 0
			for _, memoryPart := range corePart.Memory {
				currMemoryPercentageAcc += memoryPart.Percentage
			}
			if currMemoryPercentageAcc != 100 {
				return fmt.Errorf("memory partitions of <Power: %d, Cores: %d> percentages must sum 100", powerPart.Value, corePart.Value)
			}
		}
		if currCoresPercentageAcc != 100 {
			return fmt.Errorf("core partitions of <Power: %d> percentages must sum 100", powerPart.Value)
		}
	}
	if powerPercentageAcc != 100 {
		return fmt.Errorf("cpu power partitions percentages must sum 100")
	}

	// ======================= Offering Chord Discovery Backend specific ==========================

	if c.MaxRefreshesFailed() < 0 {
		return fmt.Errorf("maximum number of failed refreshes must be a positive integer")
	}

	if c.MaxRefreshesMissed() < 0 {
		return fmt.Errorf("maximum number of missed refreshes must be a positive integer")
	}

	if c.GUIDEstimatedNetworkSize() <= 0 {
		return fmt.Errorf("estimated network size must a positive integer")
	}

	if c.GUIDScaleFactor() <= 0 {
		return fmt.Errorf("GUID scale factor must a positive integer")
	}

	// ======================= Random Chord Discovery Backend specific ==========================

	if c.RandBackendMaxRetries() <= 0 {
		return fmt.Errorf("random discovery backend maximum retries must be a positive integer")
	}

	// ===================================== Overlay ============================================

	if !util.IsValidPort(c.OverlayPort()) {
		return fmt.Errorf("invalid overlay port: %d", c.OverlayPort())
	}

	// =================================== Chord Overlay =========================================

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

// Print the current configurations in order to debug the programs behavior.
func (c *Configuration) Print() {
	log.Printf("##################################################################")
	log.Printf("#                    CARAVELA's CONFIGURATIONS                   #")
	log.Printf("##################################################################")

	log.Printf("$$$$$$$$$$$$$$$$$$$$$$$$$$$ HOST $$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
	log.Printf("IP Address:                  %s", c.HostIP())
	log.Printf("Docker Engine API Version:   %s", c.DockerAPIVersion())

	log.Printf("$$$$$$$$$$$$$$$$$$$$$$$$$$ CARAVELA $$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
	log.Printf("Simulation:                  %t", c.Simulation())
	log.Printf("Port:                        %d", c.APIPort())
	log.Printf("Messages Timeout:            %s", c.APITimeout().String())
	log.Printf("CPU Slices:                  %d", c.CPUSlices())
	log.Printf("CPU Overcommit:              %d", c.CPUOvercommit())
	log.Printf("Memory Overcommit:           %d", c.MemoryOvercommit())
	log.Printf("Scheduling Policy:           %s", c.SchedulingPolicy())
	log.Printf("FreeResources Partitions:")
	for _, powerPart := range c.Caravela.Resources.CPUClasses {
		log.Printf("  CPUClass:                  %d", powerPart.Value)
		for _, corePart := range powerPart.CPUCores {
			log.Printf("    CPUCores:                %d", corePart.Value)
			for _, memoryPart := range corePart.Memory {
				log.Printf("      Memory:                %d", memoryPart.Value)
			}
		}
	}

	log.Printf("Discovery Backends:")
	log.Printf("  Active Storage Backend:            %s", c.DiscoveryBackend())
	log.Printf("    Chord-Random:")
	log.Printf("      Request Retries:               %d", c.RandBackendMaxRetries())
	log.Printf("    Chord-Offering:")
	log.Printf("      Supply FreeResources Interval: %s", c.SupplyingInterval().String())
	log.Printf("      Spread Offers Interval:        %s", c.SpreadOffersInterval().String())
	log.Printf("      Refreshes Check Interval:      %s", c.RefreshesCheckInterval().String())
	log.Printf("      Refreshes Interval:            %s", c.RefreshingInterval().String())
	log.Printf("      Refresh missed timeout:        %s", c.RefreshMissedTimeout().String())
	log.Printf("      Max num of refreshes failed:   %d", c.MaxRefreshesFailed())
	log.Printf("      Max num of refreshes missed:   %d", c.MaxRefreshesMissed())
	// Debug performance flags.
	log.Printf("      Spread Offers:                 %t", c.SpreadOffers())
	log.Printf("      Spread Partitions State:       %t", c.SpreadPartitionsState())
	log.Printf("      GUID Estimated Network Size:   %d", c.GUIDEstimatedNetworkSize())
	log.Printf("      GUID Scale Factor:             %d", c.GUIDScaleFactor())

	log.Printf("$$$$$$$$$$$$$$$$$$$$$$$ IMAGES STORAGE $$$$$$$$$$$$$$$$$$$$$$$$$$$")
	log.Printf("Active Storage Backend:              %s", c.ImagesStorageBackend())

	log.Printf("$$$$$$$$$$$$$$$$$$$$$$$$$$ OVERLAY $$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$")
	log.Printf("Active Overlay:                      %s", c.OverlayName())
	log.Printf("Port:                                %d", c.OverlayPort())

	log.Printf("Chord:")
	log.Printf("  Messages Timeout:                  %s", c.ChordTimeout().String())
	log.Printf("  Number of Virtual Nodes:           %d", c.ChordVirtualNodes())
	log.Printf("  Number of Successors:              %d", c.ChordNumSuccessors())
	log.Printf("  Hash Size (bits):                  %d", c.ChordHashSizeBits())

	log.Printf("##################################################################")
}

// ============================ Host ==============================

func (c *Configuration) HostIP() string {
	return c.Host.IP
}

func (c *Configuration) DockerAPIVersion() string {
	return c.Host.DockerAPIVersion
}

// ========================== Caravela =============================

func (c *Configuration) Simulation() bool {
	return c.Caravela.Simulation
}

func (c *Configuration) APIPort() int {
	return c.Caravela.APIPort
}

func (c *Configuration) APITimeout() time.Duration {
	return c.Caravela.APITimeout.Duration
}

func (c *Configuration) ResourcesPartitions() ResourcesPartitions {
	return c.Caravela.Resources
}

func (c *Configuration) CPUSlices() int {
	return c.Caravela.CPUSlices
}

func (c *Configuration) CPUOvercommit() int {
	return c.Caravela.CPUOvercommit
}

func (c *Configuration) MemoryOvercommit() int {
	return c.Caravela.MemoryOvercommit
}

func (c *Configuration) SchedulingPolicy() string {
	return c.Caravela.SchedulingPolicy
}

// ========================== Discovery StorageBackend ================================

func (c *Configuration) DiscoveryBackend() string {
	return c.Caravela.DiscoveryBackend.Backend
}

// ================== Offering Chord Discovery Backend specific ===================

func (c *Configuration) SpreadOffersInterval() time.Duration {
	return c.Caravela.DiscoveryBackend.OfferingChordBackend.SpreadOffersInterval.Duration
}

func (c *Configuration) SupplyingInterval() time.Duration {
	return c.Caravela.DiscoveryBackend.OfferingChordBackend.SupplyingInterval.Duration
}

func (c *Configuration) RefreshesCheckInterval() time.Duration {
	return c.Caravela.DiscoveryBackend.OfferingChordBackend.RefreshesCheckInterval.Duration
}

func (c *Configuration) RefreshingInterval() time.Duration {
	return c.Caravela.DiscoveryBackend.OfferingChordBackend.RefreshingInterval.Duration
}

func (c *Configuration) MaxRefreshesMissed() int {
	return c.Caravela.DiscoveryBackend.OfferingChordBackend.MaxRefreshesMissed
}

func (c *Configuration) MaxRefreshesFailed() int {
	return c.Caravela.DiscoveryBackend.OfferingChordBackend.MaxRefreshesFailed
}

func (c *Configuration) RefreshMissedTimeout() time.Duration {
	return c.Caravela.DiscoveryBackend.OfferingChordBackend.RefreshMissedTimeout.Duration
}

// Debug performance flag.
func (c *Configuration) SpreadOffers() bool {
	return c.Caravela.DiscoveryBackend.OfferingChordBackend.SpreadOffers
}

// Debug performance flag.
func (c *Configuration) SpreadPartitionsState() bool {
	return c.Caravela.DiscoveryBackend.OfferingChordBackend.SpreadPartitionsState
}

// Debug performance flag.
func (c *Configuration) GUIDEstimatedNetworkSize() int {
	return c.Caravela.DiscoveryBackend.OfferingChordBackend.GUIDEstimatedNetworkSize
}

// Debug performance flag.
func (c *Configuration) GUIDScaleFactor() int {
	return c.Caravela.DiscoveryBackend.OfferingChordBackend.GUIDScaleFactor
}

// =================== Random Chord Discovery Backend specific ==============

func (c *Configuration) RandBackendMaxRetries() int {
	return c.Caravela.DiscoveryBackend.RandomChordBackend.RandBackendMaxRetries
}

// ========================= Images Storage Backend =========================

func (c *Configuration) ImagesStorageBackend() string {
	return c.ImagesStorage.StorageBackend
}

// =============================== Overlay ==================================

func (c *Configuration) OverlayName() string {
	return c.Overlay.Overlay
}

func (c *Configuration) OverlayPort() int {
	return c.Overlay.OverlayPort
}

// =========================== Chord's Specific =============================

func (c *Configuration) ChordTimeout() time.Duration {
	return c.Overlay.Chord.Timeout.Duration
}

func (c *Configuration) ChordVirtualNodes() int {
	return c.Overlay.Chord.VirtualNodes
}

func (c *Configuration) ChordNumSuccessors() int {
	return c.Overlay.Chord.NumSuccessors
}

func (c *Configuration) ChordHashSizeBits() int {
	return c.Overlay.Chord.HashSizeBits
}
