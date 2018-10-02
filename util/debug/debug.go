package debug

import (
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
	"unsafe"
)

// SizeofString returns the size of the string in bytes.
func SizeofString(s string) uintptr {
	if s == "" {
		return 0
	}
	return uintptr(len(s))
}

// SizeofStringSlice returns the size of the string in bytes.
func SizeofStringSlice(slice []string) uintptr {
	sizeAccBytes := uintptr(0)
	for _, s := range slice {
		sizeAccBytes += SizeofString(s)
	}
	return sizeAccBytes
}

func SizeofPortMappings(portMappings []types.PortMapping) uintptr {
	portMappingsSizeBytes := uintptr(0)
	for i := range portMappings {
		portMappingsSizeBytes += SizeofPortMapping(&portMappings[i])
	}
	return portMappingsSizeBytes
}

func SizeofPortMapping(portMapping *types.PortMapping) uintptr {
	portMappingSizeBytes := unsafe.Sizeof(*portMapping)
	portMappingSizeBytes += SizeofString(portMapping.Protocol)
	return portMappingSizeBytes
}

func SizeofResources(res *resources.Resources) uintptr {
	guidSizeBytes := unsafe.Sizeof(*res)
	return guidSizeBytes
}

func SizeofGUID(guid *guid.GUID) uintptr {
	guidSizeBytes := unsafe.Sizeof(*guid)
	guidSizeBytes += unsafe.Sizeof(*guid.BigInt())
	return guidSizeBytes
}
