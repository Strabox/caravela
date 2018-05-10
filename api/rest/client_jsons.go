package rest

/* =================================================================================
									Request Messages
   ================================================================================= */
/*
Run container struct/JSON used in local REST APIs when a user submit a container o run
*/
type RunContainerJSON struct {
	ContainerImage string   `json:"ContainerImage"` // Container's image key
	Arguments      []string `json:"Arguments"`      // Arguments for container run
	CPUs           int      `json:"CPUs"`           // Amount of CPUs necessary
	RAM            int      `json:"RAM"`            // Amount of RAM necessary
}
