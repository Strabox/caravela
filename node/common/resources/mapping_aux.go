package resources

// Partition of a given resource through a percentage.
type ResourcePartition struct {
	Value      float64
	Percentage int
}

type ResourcePartitions struct {
	cpuPowerPartitions []CPUPowerPartition
}

func (r *ResourcePartitions) CPUPowerPartitions() []ResourcePartition {
	res := make([]ResourcePartition, 0)
	for _, partition := range r.cpuPowerPartitions {
		res = append(res, partition.ResourcePartition)
	}
	return res
}

func (r *ResourcePartitions) CPUPowerPercentages() []int {
	res := make([]int, 0)
	for _, partition := range r.cpuPowerPartitions {
		res = append(res, partition.ResourcePartition.Percentage)
	}
	return res
}

type CPUPowerPartition struct {
	ResourcePartition
	cpuCoresPartitions []CPUCoresPartition
}

func (c *CPUPowerPartition) CPUCoresPartitions() []ResourcePartition {
	res := make([]ResourcePartition, 0)
	for _, partition := range c.cpuCoresPartitions {
		res = append(res, partition.ResourcePartition)
	}
	return res
}

func (c *CPUPowerPartition) CPUCoresPercentages() []int {
	res := make([]int, 0)
	for _, partition := range c.cpuCoresPartitions {
		res = append(res, partition.ResourcePartition.Percentage)
	}
	return res
}

type CPUCoresPartition struct {
	ResourcePartition
	ramPartitions []RAMPartition
}

func (c *CPUCoresPartition) RAMPartitions() []ResourcePartition {
	res := make([]ResourcePartition, 0)
	for _, partition := range c.ramPartitions {
		res = append(res, partition.ResourcePartition)
	}
	return res
}

func (c *CPUCoresPartition) RAMPercentages() []int {
	res := make([]int, 0)
	for _, partition := range c.ramPartitions {
		res = append(res, partition.ResourcePartition.Percentage)
	}
	return res
}

type RAMPartition struct {
	ResourcePartition
}
