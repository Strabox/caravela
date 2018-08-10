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

type ResourcesPartitionV2 struct {
	Value      int `json:"Value"`
	Percentage int `json:"Percentage"`
}

type ResourcePartitionsV2 struct {
	CPUPowers []CPUPowerPartitionV2 `json:"CPUPowers"`
}

type CPUPowerPartitionV2 struct {
	ResourcesPartitionV2 `json:"ResourcesPartitionV2"`
	CPUCores             []CPUCoresPartitionV2 `json:"CPUCoresPartitions"`
}

type CPUCoresPartitionV2 struct {
	ResourcesPartitionV2 `json:"ResourcesPartitionV2"`
	RAMs                 []RAMPartitionV2 `json:"RAMPartitions"`
}

type RAMPartitionV2 struct {
	ResourcesPartitionV2 `json:"ResourcesPartitionV2"`
}
