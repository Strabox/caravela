package resources

// Partition of a given resource through a percentage.
type ResourcePartitionV2 struct {
	Value      float64
	Percentage int
}

type ResourcePartitions struct {
	cpuPowerPartitions []CPUPowerPartition
}

func (r *ResourcePartitions) CPUPowerPartitions() []ResourcePartitionV2 {
	res := make([]ResourcePartitionV2, 0)
	for _, partition := range r.cpuPowerPartitions {
		res = append(res, partition.ResourcePartitionV2)
	}
	return res
}

func (r *ResourcePartitions) CPUPowerPercentages() []int {
	res := make([]int, 0)
	for _, partition := range r.cpuPowerPartitions {
		res = append(res, partition.ResourcePartitionV2.Percentage)
	}
	return res
}

type CPUPowerPartition struct {
	ResourcePartitionV2
	cpuCoresPartitions []CPUCoresPartition
}

func (c *CPUPowerPartition) CPUCoresPartitions() []ResourcePartitionV2 {
	res := make([]ResourcePartitionV2, 0)
	for _, partition := range c.cpuCoresPartitions {
		res = append(res, partition.ResourcePartitionV2)
	}
	return res
}

func (c *CPUPowerPartition) CPUCoresPercentages() []int {
	res := make([]int, 0)
	for _, partition := range c.cpuCoresPartitions {
		res = append(res, partition.ResourcePartitionV2.Percentage)
	}
	return res
}

type CPUCoresPartition struct {
	ResourcePartitionV2
	ramPartitions []RAMPartition
}

func (c *CPUCoresPartition) RAMPartitions() []ResourcePartitionV2 {
	res := make([]ResourcePartitionV2, 0)
	for _, partition := range c.ramPartitions {
		res = append(res, partition.ResourcePartitionV2)
	}
	return res
}

func (c *CPUCoresPartition) RAMPercentages() []int {
	res := make([]int, 0)
	for _, partition := range c.ramPartitions {
		res = append(res, partition.ResourcePartitionV2.Percentage)
	}
	return res
}

type RAMPartition struct {
	ResourcePartitionV2
}
