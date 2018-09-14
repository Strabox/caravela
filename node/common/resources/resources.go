package resources

import "fmt"

// FreeResources represent of the resources that a user can ask for a container to have available.
type Resources struct {
	cpuClass int
	cpus     int
	memory   int
}

// NewResourcesCPUClass creates a new resource combination object.
func NewResourcesCPUClass(cpuClass int, cpus int, memory int) *Resources {
	return &Resources{
		cpuClass: cpuClass,
		cpus:     cpus,
		memory:   memory,
	}
}

//
func NewResources(cpus int, memory int) *Resources {
	return &Resources{
		cpuClass: 0,
		cpus:     cpus,
		memory:   memory,
	}
}

// AddCPUs adds a given number of cpus to the resources.
func (r *Resources) AddCPUs(cpus int) {
	r.cpus += cpus
}

// AddMemory adds a given amount of memory to the resources.
func (r *Resources) AddMemory(memory int) {
	r.memory += memory
}

// Add adds a given combination of resources to the receiver.
func (r *Resources) Add(resources Resources) {
	r.cpus += resources.CPUs()
	r.memory += resources.Memory()
}

// Sub subtracts a given combination of resources to the receiver.
func (r *Resources) Sub(resources Resources) {
	r.cpus -= resources.CPUs()
	r.memory -= resources.Memory()
}

// SetZero sets the resources to zero.
func (r *Resources) SetZero() {
	r.cpus = 0
	r.memory = 0
}

// SetTo sets the resources into a specific combination of resources.
func (r *Resources) SetTo(resources Resources) {
	r.cpuClass = resources.CPUClass()
	r.cpus = resources.CPUs()
	r.memory = resources.Memory()
}

// IsZero returns true if the resources are zero, false otherwise.
func (r *Resources) IsZero() bool {
	return r.cpus == 0 && r.memory == 0
}

// IsValid return true if all resources are greater than 0.
func (r *Resources) IsValid() bool {
	return r.cpus > 0 && r.memory > 0
}

// IsNegative return true if one of the resources has a negative amount.
func (r *Resources) IsNegative() bool {
	return r.cpus < 0 || r.memory < 0
}

// Contains returns true if the given resources are contained inside the receiver.
func (r *Resources) Contains(contained Resources) bool {
	return r.CPUClass() >= contained.CPUClass() && r.cpus >= contained.CPUs() && r.memory >= contained.Memory()
}

// Equals returns true if the given resource combination is equal to the receiver.
func (r *Resources) Equals(resources Resources) bool {
	return r.cpuClass == resources.cpuClass && r.cpus == resources.cpus && r.memory == resources.memory
}

// Copy returns a object that is a exact copy of the receiver.
func (r *Resources) Copy() *Resources {
	return &Resources{
		cpuClass: r.cpuClass,
		cpus:     r.cpus,
		memory:   r.memory,
	}
}

// String stringify the receiver resources object.
func (r *Resources) String() string {
	return fmt.Sprintf("<<%d;%d>;%d>", r.cpuClass, r.cpus, r.memory)
}

// CPUClass getter.
func (r *Resources) CPUClass() int {
	return r.cpuClass
}

// CPUs getter.
func (r *Resources) CPUs() int {
	return r.cpus
}

// Memory getter.
func (r *Resources) Memory() int {
	return r.memory
}

// SetCPUClass CPU Class setter.
func (r *Resources) SetCPUClass(cpuClass int) {
	r.cpuClass = cpuClass
}

// SetCPUs CPUs setter.
func (r *Resources) SetCPUs(cpu int) {
	r.cpus = cpu
}

// SetMemory Memory setter.
func (r *Resources) SetMemory(memory int) {
	r.memory = memory
}
