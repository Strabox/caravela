package util

import (
	"github.com/shirou/gopsutil/cpu"
)

// GetCpuClass ...
func GetCpuClass() int {
	for i := 0; i < 3; i++ {
		cpus, err := cpu.Info()
		if err == nil && len(cpus) > 0 {
			if cpus[0].Family == "198" { // x86?
				if cpus[0].Mhz < 2000 {
					return 0
				} else {
					return 1
				}
			}
		}
	}
	return 0 // If we can't get the CPU class we return the lowest one.
}
