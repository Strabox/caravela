package resources

import (
	"fmt"
)

/*
Representation of the resources that a user can ask for a container to have available.
*/
type Resources struct {
	vCPU int
	ram  int
}

func NewResources(vCPU int, RAM int) *Resources {
	return &Resources{vCPU, RAM}
}

func (r *Resources) CPU() int {
	return r.vCPU
}

func (r *Resources) RAM() int {
	return r.ram
}

func (r *Resources) SetCPU(cpu int) {
	r.vCPU = cpu
}

func (r *Resources) SetRAM(ram int) {
	r.ram = ram
}

func (r *Resources) AddCPU(cpu int) {
	r.vCPU += cpu
}

func (r *Resources) AddRAM(ram int) {
	r.ram += ram
}

func (r *Resources) Add(resources Resources) {
	r.vCPU += resources.CPU()
	r.ram += resources.RAM()
}

func (r *Resources) SetZero() {
	r.vCPU = 0
	r.ram = 0
}

func (r *Resources) SetTo(resources Resources) {
	r.vCPU = resources.CPU()
	r.ram = resources.RAM()
}

func (r *Resources) IsZero() bool {
	return r.vCPU == 0 && r.ram == 0
}

func (r *Resources) Contains(r2 Resources) bool {
	return r.vCPU >= r2.CPU() && r.ram >= r2.RAM()
}

func (r *Resources) Equals(r2 Resources) bool {
	return r.vCPU == r2.vCPU && r.ram == r2.ram
}

func (r *Resources) Copy() *Resources {
	res := &Resources{}
	res.vCPU = r.vCPU
	res.ram = r.ram
	return res
}

func (r *Resources) String() string {
	return fmt.Sprintf("CPUs: %d RAM: %d", r.vCPU, r.ram)
}
