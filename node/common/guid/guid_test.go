package guid

import (
	"github.com/stretchr/testify/assert"
	"math/big"
	"strconv"
	"testing"
)

func TestInitializeGUID(t *testing.T) {
	Init(256)

	assert.Equal(t, 256, SizeBits(), "GUID size were not correctly set!")
	assert.Equal(t, 256/8, SizeBytes(), "GUID size were not correctly set!")
}

func TestInitializeGUID_Reinitialize(t *testing.T) {
	Init(256)
	Init(160)

	assert.Equal(t, 256, SizeBits(), "GUID size were reinitialized after the first time!")
}

func TestGuidSizeBits(t *testing.T) {
	assert.Equal(t, 256, SizeBits(), "GUID size bits returned wrong value!")
}

func TestGuidSizeBytes(t *testing.T) {
	assert.Equal(t, 256/8, SizeBytes(), "GUID size bytes returned wrong value!")
}

func TestNewZero(t *testing.T) {
	zeroGUID := NewZero()

	assert.Equal(t, "0", zeroGUID.String(), "Zero GUID value is wrong!")
	assert.Equal(t, int64(0), zeroGUID.Int64(), "Zero GUID value is wrong!")
	assert.Equal(t, SizeBytes(), len(zeroGUID.Bytes()), "Byte array return has different length from the GUID size")
}

func TestMaximumGUID(t *testing.T) {
	expectedMax := big.NewInt(0)
	expectedMax.Exp(big.NewInt(2), big.NewInt(int64(SizeBits())), nil)
	expectedMax = expectedMax.Sub(expectedMax, big.NewInt(1))

	max := MaximumGUID()

	assert.Equal(t, expectedMax.String(), max.String(), "Maximum GUID value is wrong!")
	assert.Equal(t, SizeBytes(), len(max.Bytes()), "Byte array return has different length from the GUID size")
}

func TestNewGUIDRandom(t *testing.T) {
	randGUID := NewGUIDRandom()

	assert.True(t, randGUID.Lower(*MaximumGUID()), "GUID should be smaller than MAX GUID")
	assert.True(t, randGUID.Higher(*NewGUIDInteger(0)), "GUID should be higher than 0")
}

func TestNewGUIDString(t *testing.T) {
	guidString := "100000"

	guid := NewGUIDString(guidString)

	assert.True(t, guid.Equals(*NewGUIDInteger(100000)), "GUID object created was wrongly created")
	assert.Equal(t, guidString, guid.String(), "GUID object created was wrongly created")
	assert.Equal(t, SizeBytes(), len(guid.Bytes()), "Byte array return has different length from the GUID size")
}

func TestNewGUIDInteger(t *testing.T) {
	guidInteger := int64(10000)

	guid := NewGUIDInteger(guidInteger)

	assert.True(t, guid.Equals(*NewGUIDInteger(10000)), "GUID object created was wrongly created")
	assert.Equal(t, strconv.Itoa(int(guidInteger)), guid.String(), "GUID object created was wrongly created")
	assert.Equal(t, SizeBytes(), len(guid.Bytes()), "Byte array return has different length from the GUID size")
}

func TestNewGUIDBytes(t *testing.T) {
	guidBytes := []byte{0, 0, 0, 0, 0, 0, 0, 0, 1}
	guidBytesExpected := make([]byte, SizeBytes())
	guidBytesExpected[SizeBytes()-1] = 1

	guid := NewGUIDBytes(guidBytes)

	assert.True(t, guid.Equals(*NewGUIDInteger(1)), "GUID object created was wrongly created")
	assert.Equal(t, guidBytesExpected, guid.Bytes(), "GUID object created was wrongly created")
	assert.Equal(t, SizeBytes(), len(guid.Bytes()), "Byte array return has different length from the GUID size")
}

func TestNewGUIDBigInt(t *testing.T) {
	guidBigInt := big.NewInt(7899999999999)

	guid := newGUIDBigInt(guidBigInt)

	assert.True(t, guid.Equals(*NewGUIDInteger(7899999999999)), "GUID object created was wrongly created")
	assert.Equal(t, strconv.Itoa(int(7899999999999)), guid.String(), "GUID object created was wrongly created")
	assert.Equal(t, SizeBytes(), len(guid.Bytes()), "Byte array return has different length from the GUID size")
}

func TestGuid_GenerateRandom_BetweenTwoEqualGUIDs(t *testing.T) {
	lowerBound := int64(5)
	higherBound := int64(5)
	lowerGUID := NewGUIDInteger(lowerBound)
	higherGUID := NewGUIDInteger(higherBound)

	// Hack: Generate several randoms to check in order to improve our statistical success :)
	for i := 0; i < 50; i++ {
		randGUID, _ := lowerGUID.GenerateInnerRandomGUID(*higherGUID)

		assert.True(t, randGUID.Equals(*NewGUIDInteger(5)), "Random GUID is not in acceptable range")
		assert.Equal(t, "5", randGUID.String(), "Random GUID is not in acceptable range")
		assert.Equal(t, SizeBytes(), len(randGUID.Bytes()), "Byte array return has different length from the GUID size")
	}
}

func TestGuid_GenerateRandom_BetweenTwoDifferentGUIDs(t *testing.T) {
	lowerBound := int64(0)
	higherBound := int64(5)
	lowerGUID := NewGUIDInteger(lowerBound)
	higherGUID := NewGUIDInteger(higherBound)
	// Set of acceptable outputs of the random GUID generation
	acceptableOutputs := make(map[int64]interface{})
	for i := lowerBound; i < higherBound; i++ {
		acceptableOutputs[i] = nil
	}

	// Hack: Generate several randoms to check in order to improve our statistical success :)
	for i := 0; i < 100000; i++ {
		randGUID, _ := lowerGUID.GenerateInnerRandomGUID(*higherGUID)

		_, exist := acceptableOutputs[randGUID.Int64()]
		assert.True(t, exist, "Random GUID is not in acceptable range")
		assert.Equal(t, SizeBytes(), len(randGUID.Bytes()), "Byte array return has different length from the GUID size")
	}
}

func TestGuid_GenerateRandom_BetweenTwoDifferentGUIDsConcurrently(t *testing.T) {
	lowerBound := int64(0)
	higherBound := int64(5)
	lowerGUID := NewGUIDInteger(lowerBound)
	higherGUID := NewGUIDInteger(higherBound)
	// Set of acceptable outputs of the random GUID generation
	acceptableOutputs := make(map[int64]interface{})
	for i := lowerBound; i < higherBound; i++ {
		acceptableOutputs[i] = nil
	}

	// Hack: Generate several randoms to check in order to improve our statistical success :)
	for i := 0; i < 100000; i++ {
		// Use multiple goroutines to check if it is safe.
		go func() {
			randGUID, _ := lowerGUID.GenerateInnerRandomGUID(*higherGUID)

			_, exist := acceptableOutputs[randGUID.Int64()]
			assert.True(t, exist, "Random GUID is not in acceptable range")
			assert.Equal(t, SizeBytes(), len(randGUID.Bytes()), "Byte array return has different length from the GUID size")
		}()
	}
}

func TestGuid_PercentageOffset(t *testing.T) {
	lowerBound := int64(100)
	higherBound := int64(1000)
	lowerGUID := NewGUIDInteger(lowerBound)
	higherGUID := NewGUIDInteger(higherBound)

	offset := lowerGUID.PercentageOffset(50, *higherGUID)

	assert.Equal(t, offset, "450", "Wrong percentage offset")
}

func TestGuid_PercentageOffset_EqualGUIDs(t *testing.T) {
	lowerBound := int64(100)
	higherBound := int64(100)
	lowerGUID := NewGUIDInteger(lowerBound)
	higherGUID := NewGUIDInteger(higherBound)

	offset := lowerGUID.PercentageOffset(50, *higherGUID)

	assert.Equal(t, offset, "0", "Wrong percentage offset")
}

func TestGuid_PercentageOffset_EqualOffset0(t *testing.T) {
	lowerBound := int64(100)
	higherBound := int64(1000)
	lowerGUID := NewGUIDInteger(lowerBound)
	higherGUID := NewGUIDInteger(higherBound)

	offset := lowerGUID.PercentageOffset(0, *higherGUID)

	assert.Equal(t, offset, "0", "Wrong percentage offset")
}

func TestGuid_PercentageOffset_EqualOffset100(t *testing.T) {
	lowerBound := int64(100)
	higherBound := int64(1000)
	lowerGUID := NewGUIDInteger(lowerBound)
	higherGUID := NewGUIDInteger(higherBound)

	offset := lowerGUID.PercentageOffset(100, *higherGUID)

	assert.Equal(t, offset, "900", "Wrong percentage offset")
}

func TestGuid_AddOffset_Positive(t *testing.T) {
	stringOffset := "100"
	guid := NewGUIDInteger(56)

	guid.AddOffset(stringOffset)

	assert.True(t, guid.Equals(*NewGUIDInteger(156)), "Wrong GUID value")
	assert.Equal(t, "156", guid.String(), "Wrong GUID value")
	assert.Equal(t, SizeBytes(), len(guid.Bytes()), "Byte array return has different length from the GUID")
}

func TestGuid_Cmp_Equal(t *testing.T) {
	guid1 := NewGUIDInteger(67)
	guid2 := NewGUIDInteger(67)

	res := guid1.Cmp(*guid2)

	assert.Equal(t, 0, res, "GUIDs should be equal")
}

func TestGuid_Cmp_Greater(t *testing.T) {
	guid1 := NewGUIDInteger(68)
	guid2 := NewGUIDInteger(67)

	res := guid1.Cmp(*guid2)

	assert.Equal(t, 1, res, "Receiver should be greater")
}

func TestGuid_Cmp_Lower(t *testing.T) {
	guid1 := NewGUIDInteger(66)
	guid2 := NewGUIDInteger(67)

	res := guid1.Cmp(*guid2)

	assert.Equal(t, -1, res, "Receiver should be lower")
}

func TestGuid_Higher_True(t *testing.T) {
	guid1 := NewGUIDInteger(68)
	guid2 := NewGUIDInteger(67)

	res := guid1.Higher(*guid2)

	assert.True(t, res, "Receiver should be higher")
}

func TestGuid_Higher_False(t *testing.T) {
	guid1 := NewGUIDInteger(888)
	guid2 := NewGUIDInteger(887)

	res := guid1.Lower(*guid2)

	assert.False(t, res, "Receiver should be lower")
}

func TestGuid_Lower_True(t *testing.T) {
	guid1 := NewGUIDInteger(999)
	guid2 := NewGUIDInteger(1000)

	res := guid1.Lower(*guid2)

	assert.True(t, res, "Receiver should be lower")
}

func TestGuid_Lower_False(t *testing.T) {
	guid1 := NewGUIDInteger(2)
	guid2 := NewGUIDInteger(1)

	res := guid1.Lower(*guid2)

	assert.False(t, res, "Receiver should be lower")
}

func TestGuid_Equals_True(t *testing.T) {
	guid1 := NewGUIDInteger(999)
	guid2 := NewGUIDInteger(999)

	equal := guid1.Equals(*guid2)

	assert.True(t, equal, "GUIDs should be equal")
}

func TestGuid_Equals_False_1(t *testing.T) {
	guid1 := NewGUIDInteger(999)
	guid2 := NewGUIDInteger(998)

	equal := guid1.Equals(*guid2)

	assert.False(t, equal, "GUIDs should be equal")
}

func TestGuid_Equals_False_2(t *testing.T) {
	guid1 := NewGUIDInteger(999)
	guid2 := NewGUIDInteger(1000)

	equal := guid1.Equals(*guid2)

	assert.False(t, equal, "GUIDs should be equal")
}

func TestGuid_Short(t *testing.T) {
	const GUID = "77546568768548746546877"

	guid := NewGUIDString(GUID)
	shortGUID := guid.Short()

	assert.Equal(t, guid.String()[0:guidShortStringSize], shortGUID, "Byte array return has different length from the GUID")
}

func TestGuid_Copy(t *testing.T) {
	guidOriginal := NewGUIDString("7777")

	guidCopy := guidOriginal.Copy()

	assert.True(t, guidOriginal.Equals(*guidCopy), "GUIDs should be equal")
	assert.Equal(t, guidOriginal.String(), guidCopy.String(), "Strings should be equal")
	assert.Equal(t, SizeBytes(), len(guidCopy.Bytes()), "Byte array return has different length from the GUID")
}
