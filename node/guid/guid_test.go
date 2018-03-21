package guid

import (
	"github.com/stretchr/testify/assert"
	"math/big"
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

func TestGuidSizeBits(t *testing.T) {

	assert.Equal(t, GuidSizeBits(), 256, "Guid size bits returned wrong value!")
}

func TestGuidSizeBytes(t *testing.T) {

	assert.Equal(t, GuidSizeBytes(), 256/8, "Guid size bytes returned wrong value!")
}

func TestMaximumGuid(t *testing.T) {
	test := big.NewInt(0)
	test.Exp(big.NewInt(2), big.NewInt(int64(GuidSizeBits())), nil)
	test = test.Sub(test, big.NewInt(1))

	max := MaximumGuid()

	assert.Equal(t, max.String(), test.String(), "Maximum GUID value is wrong!")
}
