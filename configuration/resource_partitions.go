package configuration

// Represents a configuration partition in terms of percentage,
type ResourcesPartition struct {
	Percentage int
}

// Partition for CPU power.
type CPUPowerPartition struct {
	ResourcesPartition
	Class int
}

// Partition for the number of CPU cores.
type CPUCoresPartition struct {
	ResourcesPartition
	Cores int
}

// Partition for the amount of RAM.
type RAMPartition struct {
	ResourcesPartition
	RAM int
}
