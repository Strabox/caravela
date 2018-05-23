package configuration

/*
Partition for cpu power.
*/
type CpuPowerPartition struct {
	Class      int
	Percentage int
}

/*
Partition for the number of cpu cores.
*/
type CpuCoresPartition struct {
	Cores      int
	Percentage int
}

/*
Partition for the amount of ram.
*/
type RamPartition struct {
	Ram        int
	Percentage int
}
