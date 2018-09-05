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
	memoryPartitions []MemoryPartition
}

func (c *CPUCoresPartition) MemoryPartitions() []ResourcePartition {
	res := make([]ResourcePartition, 0)
	for _, partition := range c.memoryPartitions {
		res = append(res, partition.ResourcePartition)
	}
	return res
}

func (c *CPUCoresPartition) MemoryPercentages() []int {
	res := make([]int, 0)
	for _, partition := range c.memoryPartitions {
		res = append(res, partition.ResourcePartition.Percentage)
	}
	return res
}

type MemoryPartition struct {
	ResourcePartition
}
