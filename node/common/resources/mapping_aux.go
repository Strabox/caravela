package resources

// Partition of a given resource through a percentage.
type ResourcePartition struct {
	Value      float64
	Percentage int
}

type ResourcePartitions struct {
	cpuClassPartitions []CPUClassPartition
}

func (r *ResourcePartitions) CPUClassPartitions() []ResourcePartition {
	res := make([]ResourcePartition, 0)
	for _, partition := range r.cpuClassPartitions {
		res = append(res, partition.ResourcePartition)
	}
	return res
}

func (r *ResourcePartitions) CPUClassPercentages() []int {
	res := make([]int, 0)
	for _, partition := range r.cpuClassPartitions {
		res = append(res, partition.ResourcePartition.Percentage)
	}
	return res
}

type CPUClassPartition struct {
	ResourcePartition
	cpuCoresPartitions []CPUCoresPartition
}

func (c *CPUClassPartition) CPUCoresPartitions() []ResourcePartition {
	res := make([]ResourcePartition, 0)
	for _, partition := range c.cpuCoresPartitions {
		res = append(res, partition.ResourcePartition)
	}
	return res
}

func (c *CPUClassPartition) CPUCoresPercentages() []int {
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
