package guid

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInitializeGuid(t *testing.T) {

	InitializeGuid(256)

	assert.Equal(t, GuidSizeBits(), 256, "Guid size were not correctly set!")
}

func TestTryReinitializeGuid(t *testing.T) {

	InitializeGuid(256)
	InitializeGuid(160)

	assert.Equal(t, GuidSizeBits(), 256, "Guid size were reinitialized after the first time!")
}
