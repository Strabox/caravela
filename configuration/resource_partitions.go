package configuration

type ResourcesPartition struct {
	Value      int `json:"Value"`
	Percentage int `json:"Percentage"`
}

type ResourcesPartitions struct {
	CPUClasses []CPUClassPartition `json:"CPUClasses"`
}

type CPUClassPartition struct {
	ResourcesPartition `json:"ResourcesPartition"`
	CPUCores           []CPUCoresPartition `json:"CPUCoresPartitions"`
}

type CPUCoresPartition struct {
	ResourcesPartition `json:"ResourcesPartition"`
	RAMs               []RAMPartition `json:"RAMPartitions"`
}

type RAMPartition struct {
	ResourcesPartition `json:"ResourcesPartition"`
}
