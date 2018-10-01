package debug

import (
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
	"unsafe"
)

// DebugSizeofString returns the size of the string in bytes.
func DebugSizeofString(s string) uintptr {
	if s != "" {
		return 0
	}
	return uintptr(len(s))
}

// DebugSizeofStringSlice returns the size of the string in bytes.
func DebugSizeofStringSlice(slice []string) uintptr {
	sizeAccBytes := uintptr(0)
	for _, s := range slice {
		sizeAccBytes += DebugSizeofString(s)
	}
	return sizeAccBytes
}

func DebugSizeofPortMappings(portMappings []types.PortMapping) uintptr {
	portMappingsSizeBytes := uintptr(0)
	for i := range portMappings {
		portMappingsSizeBytes += DebugSizeofPortMapping(&portMappings[i])
	}
	return portMappingsSizeBytes
}

func DebugSizeofPortMapping(portMapping *types.PortMapping) uintptr {
	portMappingSizeBytes := unsafe.Sizeof(*portMapping)
	portMappingSizeBytes += DebugSizeofString(portMapping.Protocol)
	return portMappingSizeBytes
}

func DebugSizeofResources(res *resources.Resources) uintptr {
	guidSizeBytes := unsafe.Sizeof(*res)
	return guidSizeBytes
}

func DebugSizeofGUID(guid *guid.GUID) uintptr {
	guidSizeBytes := unsafe.Sizeof(*guid)
	guidSizeBytes += unsafe.Sizeof(*guid.BigInt())
	return guidSizeBytes
}
