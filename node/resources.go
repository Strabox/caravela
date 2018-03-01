package node

import ()

type Resources struct {
	vCPU int
	RAM  int
}

func NewResources(vCPU int, RAM int) *Resources {
	return &Resources{vCPU, RAM}
}
