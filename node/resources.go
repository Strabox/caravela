package node

import (
"strconv"
)

type Resources struct {
	vCPU int
	RAM  int
}

func NewResources(vCPU int, RAM int) *Resources {
	return &Resources{vCPU, RAM}
}

func (r *Resources) ToString() string {
	return "CPUs: " + strconv.Itoa(r.vCPU) + " RAM: " + strconv.Itoa(r.RAM); 
}
