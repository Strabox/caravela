package resources

import "fmt"

/*
Representation of the resources that a user can ask for a container to have available.
*/
type Resources struct {
	cpus int
	ram  int
}

/*
Create a new resource combination object.
*/
func NewResources(cpus int, ram int) *Resources {
	return &Resources{cpus: cpus, ram: ram}
}

/*
Adds a given number of cpus to the resources.
*/
func (r *Resources) AddCPU(cpus int) {
	r.cpus += cpus
}

/*
Adds a given amount of ram to the resources.
*/
func (r *Resources) AddRAM(ram int) {
	r.ram += ram
}

/*
Adds a given combination of resources to the receiver.
*/
func (r *Resources) Add(resources Resources) {
	r.cpus += resources.CPUs()
	r.ram += resources.RAM()
}

/*
Subtract a given combination of resources to the receiver.
*/
func (r *Resources) Sub(resources Resources) {
	r.cpus -= resources.CPUs()
	r.ram -= resources.RAM()
}

/*
Set the resources to zero.
*/
func (r *Resources) SetZero() {
	r.cpus = 0
	r.ram = 0
}

/*
Set the resources into a specific combination of resources.
*/
func (r *Resources) SetTo(resources Resources) {
	r.cpus = resources.CPUs()
	r.ram = resources.RAM()
}

/*
Return true if the resources are zero, false otherwise.
*/
func (r *Resources) IsZero() bool {
	return r.cpus == 0 && r.ram == 0
}

/*
Return true if all resources are greater than 0
*/
func (r *Resources) Available() bool {
	return r.cpus > 0 && r.ram > 0
}

/*
Returns true if the given resources are contained inside the receiver.
*/
func (r *Resources) Contains(contained Resources) bool {
	return r.cpus >= contained.CPUs() && r.ram >= contained.RAM()
}

/*
Returns true if the given resource combination is equal to the receiver.
*/
func (r *Resources) Equals(resources Resources) bool {
	return r.cpus == resources.cpus && r.ram == resources.ram
}

/*
Returns a object that is a exact copy of the receiver.
*/
func (r *Resources) Copy() *Resources {
	res := &Resources{}
	res.cpus = r.cpus
	res.ram = r.ram
	return res
}

/*
Stringify the receiver resources object.
*/
func (r *Resources) String() string {
	return fmt.Sprintf("CPUs: %d RAM: %d", r.cpus, r.ram)
}

func (r *Resources) CPUs() int {
	return r.cpus
}

func (r *Resources) RAM() int {
	return r.ram
}

func (r *Resources) SetCPUs(cpu int) {
	r.cpus = cpu
}

func (r *Resources) SetRAM(ram int) {
	r.ram = ram
}
