package rest

/* =================================================================================
									Request Messages
   ================================================================================= */

/*
RunContainer container struct/JSON used in local REST APIs when a user submit a container o run
*/
type RunContainerMessage struct {
	ContainerImageKey string        `json:"ContainerImageKey"` // Container's image key
	PortMappings      []PortMapping `json:"PortMappings"`      // Port mappings for the container
	Arguments         []string      `json:"Arguments"`         // Arguments for the container run
	CPUs              int           `json:"CPUs"`              // Amount of CPUs necessary to run the container
	RAM               int           `json:"RAM"`               // Amount of RAM necessary to run the container
}

/*
Represents a port mapping between a container port and the host port.
*/
type PortMapping struct {
	HostPort      int `json:"HostPort"`
	ContainerPort int `json:"ContainerPort"`
}

/*
Stop containers struct/JSON used in local REST APIs when a user submit a request to stop containers
*/
type StopContainersMessage struct {
	ContainersIDs []string `json:"ContainersIDs"` // ID's of the containers to be stopped
}

/*
Represents a list o all container and its status in the node.
*/
type ContainersList struct {
	ContainersStatus []ContainerStatus `json:"ContainersStatus"`
}

/*
Represents a container and its status in the node.
*/
type ContainerStatus struct {
	ID     string `json:"ID"`
	Status string `json:"Status"`
}
