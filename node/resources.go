package node

import ()

type Resources struct {
	vCPU uint
	RAM  uint
}

func NewResources(vCPU uint, RAM uint) *Resources {
	return &Resources{vCPU, RAM}
}
