package node

import ()

type ResourcesMap struct {
	resourcesIdMap  map[uint][]*Guid
	cpuCombinations uint
	ramCombinations uint
}

func NewResourcesMap(cpuCombinations uint, ramCombinations uint) *ResourcesMap {
	rm := &ResourcesMap{}
	rm.resourcesIdMap = make(map[uint][]*Guid)
	return rm
}
