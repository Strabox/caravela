package rest

import "github.com/strabox/caravela/api/types"

/* =================================================================================
									Request Messages
   ================================================================================= */

// RunContainer container struct/JSON used in local REST APIs when a user submit a container to run.
type RunContainerMsg struct {
	ContainerImageKey string              `json:"ContainerImageKey"` // Container's image key
	PortMappings      []types.PortMapping `json:"PortMappings"`      // Port mappings for the container
	Arguments         []string            `json:"Arguments"`         // Arguments for the container run
	CPUs              int                 `json:"CPUs"`              // Amount of CPUs necessary to run the container
	RAM               int                 `json:"RAM"`               // Amount of RAM necessary to run the container
}

// Stop containers struct/JSON used in local REST APIs when a user submit a request to stop containers
type StopContainersMsg struct {
	ContainersIDs []string `json:"ContainersIDs"` // ID's of the containers to be stopped
}

// Represents a list o all container and its status in the node.
type ContainersStatusMsg struct {
	ContainersStatus []types.ContainerStatus `json:"ContainersStatus"`
}
