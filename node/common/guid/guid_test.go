package guid

import (
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestInitializeGuid(t *testing.T) {

	InitializeGUID(256)

	assert.Equal(t, SizeBits(), 256, "GUID size were not correctly set!")
}

func TestTryReinitializeGuid(t *testing.T) {

	InitializeGUID(256)
	InitializeGUID(160)

	assert.Equal(t, SizeBits(), 256, "GUID size were reinitialized after the first time!")
}

func TestGuidSizeBits(t *testing.T) {

	assert.Equal(t, SizeBits(), 256, "GUID size bits returned wrong value!")
}

func TestGuidSizeBytes(t *testing.T) {

	assert.Equal(t, SizeBytes(), 256/8, "GUID size bytes returned wrong value!")
}

func TestMaximumGuid(t *testing.T) {
	test := big.NewInt(0)
	test.Exp(big.NewInt(2), big.NewInt(int64(SizeBits())), nil)
	test = test.Sub(test, big.NewInt(1))

	max := MaximumGuid()

	assert.Equal(t, max.String(), test.String(), "Maximum GUID value is wrong!")
}
