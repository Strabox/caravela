package rest

/* =================================================================================
									Request Messages
   ================================================================================= */

/*
Represents a port mapping between a container port and the host port.
*/
type PortMappingJSON struct {
	HostPort      int `json:"HostPort"`
	ContainerPort int `json:"ContainerPort"`
}

/*
Run container struct/JSON used in local REST APIs when a user submit a container o run
*/
type RunContainerJSON struct {
	ContainerImageKey string            `json:"ContainerImageKey"` // Container's image key
	PortMappings      []PortMappingJSON `json:"PortMappings"`      // Port mappings for the container
	Arguments         []string          `json:"Arguments"`         // Arguments for the container run
	CPUs              int               `json:"CPUs"`              // Amount of CPUs necessary to run the container
	RAM               int               `json:"RAM"`               // Amount of RAM necessary to run the container
}
