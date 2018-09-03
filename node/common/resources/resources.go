package resources

import "fmt"

// FreeResources represent of the resources that a user can ask for a container to have available.
type Resources struct {
	cpuClass int
	cpus     int
	ram      int
}

// NewResourcesCPUClass creates a new resource combination object.
func NewResourcesCPUClass(cpuClass int, cpus int, ram int) *Resources {
	return &Resources{
		cpuClass: cpuClass,
		cpus:     cpus,
		ram:      ram,
	}
}

//
func NewResources(cpus int, ram int) *Resources {
	return &Resources{
		cpus: cpus,
		ram:  ram,
	}
}

// AddCPUs adds a given number of cpus to the resources.
func (r *Resources) AddCPUs(cpus int) {
	r.cpus += cpus
}

// AddRAM adds a given amount of ram to the resources.
func (r *Resources) AddRAM(ram int) {
	r.ram += ram
}

// Add adds a given combination of resources to the receiver.
func (r *Resources) Add(resources Resources) {
	r.cpus += resources.CPUs()
	r.ram += resources.RAM()
}

// Sub subtracts a given combination of resources to the receiver.
func (r *Resources) Sub(resources Resources) {
	r.cpus -= resources.CPUs()
	r.ram -= resources.RAM()
}

// SetZero sets the resources to zero.
func (r *Resources) SetZero() {
	r.cpus = 0
	r.ram = 0
}

// SetTo sets the resources into a specific combination of resources.
func (r *Resources) SetTo(resources Resources) {
	r.cpuClass = resources.CPUClass()
	r.cpus = resources.CPUs()
	r.ram = resources.RAM()
}

// IsZero returns true if the resources are zero, false otherwise.
func (r *Resources) IsZero() bool {
	return r.cpus == 0 && r.ram == 0
}

// IsValid return true if all resources are greater than 0.
func (r *Resources) IsValid() bool {
	return r.cpus > 0 && r.ram > 0
}

// IsNegative return true if one of the resources has a negative amount.
func (r *Resources) IsNegative() bool {
	return r.cpus < 0 || r.ram < 0
}

// Contains returns true if the given resources are contained inside the receiver.
func (r *Resources) Contains(contained Resources) bool {
	return r.CPUClass() >= contained.CPUClass() && r.cpus >= contained.CPUs() && r.ram >= contained.RAM()
}

// Equals returns true if the given resource combination is equal to the receiver.
func (r *Resources) Equals(resources Resources) bool {
	return r.cpuClass == resources.cpuClass && r.cpus == resources.cpus && r.ram == resources.ram
}

// Copy returns a object that is a exact copy of the receiver.
func (r *Resources) Copy() *Resources {
	return &Resources{
		cpuClass: r.cpuClass,
		cpus:     r.cpus,
		ram:      r.ram,
	}
}

// String stringify the receiver resources object.
func (r *Resources) String() string {
	return fmt.Sprintf("<<%d;%d>;%d>", r.cpuClass, r.cpus, r.ram)
}

// CPUClass getter.
func (r *Resources) CPUClass() int {
	return r.cpuClass
}

// CPUs getter.
func (r *Resources) CPUs() int {
	return r.cpus
}

// RAM getter.
func (r *Resources) RAM() int {
	return r.ram
}

// SetCPUClass CPU Class setter.
func (r *Resources) SetCPUClass(cpuClass int) {
	r.cpuClass = cpuClass
}

// SetCPUs CPUs setter.
func (r *Resources) SetCPUs(cpu int) {
	r.cpus = cpu
}

// SetRAM RAM setter.
func (r *Resources) SetRAM(ram int) {
	r.ram = ram
}
