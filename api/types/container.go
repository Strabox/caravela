package types

type ContainerConfig struct {
	ImageKey     string        `json:"ImageKey"`
	Args         []string      `json:"Args"`
	PortMappings []PortMapping `json:"PortMappings"`
	Resources    Resources     `json:"Resources"`
}

type ContainerStatus struct {
	ContainerConfig `json:"ContainerConfig"`
	ContainerID     string `json:"ContainerID"`
	Status          string `json:"Status"`
}

type PortMapping struct {
	HostPort      int `json:"HostPort"`
	ContainerPort int `json:"ContainerPort"`
}
