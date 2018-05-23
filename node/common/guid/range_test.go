package guid

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewGUIDRange(t *testing.T) {
	lowerGUID := NewGUIDInteger(10000)
	higherGUID := NewGUIDInteger(25000)

	guidRange := NewGUIDRange(*lowerGUID, *higherGUID)

	assert.True(t, lowerGUID.Equals(*guidRange.LowerGUID()), "Lower GUID of the range is wrong")
	assert.True(t, higherGUID.Equals(*guidRange.HigherGUID()), "Higher GUID of the range is wrong")
}

func TestRange_GenerateRandomInside(t *testing.T) {
	lowerBound := int64(500)
	higherBound := int64(505)
	lowerGUID := NewGUIDInteger(lowerBound)
	higherGUID := NewGUIDInteger(higherBound)
	guidRange := NewGUIDRange(*lowerGUID, *higherGUID)

	// Set of acceptable outputs of the random GUID generation
	acceptableOutputs := make(map[int64]interface{})
	for i := lowerBound; i < higherBound; i++ {
		acceptableOutputs[i] = nil
	}
	// Hack: Generate several randoms to check in order to improve our statistical success :)
	for i := 0; i < 100; i++ {
		randGUID, _ := guidRange.GenerateRandomInside()

		_, exist := acceptableOutputs[randGUID.Int64()]
		assert.True(t, exist, "Random GUID is not in acceptable range")
		assert.Equal(t, SizeBytes(), len(randGUID.Bytes()), "Byte array return has different length from the GUID size")
	}
}

func TestRange_CreatePartitions(t *testing.T) {
	lowerBound := int64(500)
	higherBound := int64(700)
	lowerGUID := NewGUIDInteger(lowerBound)
	higherGUID := NewGUIDInteger(higherBound)
	guidRange := NewGUIDRange(*lowerGUID, *higherGUID)
	partitions := []int{20, 20, 20, 20, 10, 10}

	guidRanges := guidRange.CreatePartitions(partitions)

	expectedRanges := []*Range{NewGUIDRange(*NewGUIDInteger(500), *NewGUIDInteger(540)),
		NewGUIDRange(*NewGUIDInteger(540), *NewGUIDInteger(580)),
		NewGUIDRange(*NewGUIDInteger(580), *NewGUIDInteger(620)),
		NewGUIDRange(*NewGUIDInteger(620), *NewGUIDInteger(660)),
		NewGUIDRange(*NewGUIDInteger(660), *NewGUIDInteger(680)),
		NewGUIDRange(*NewGUIDInteger(680), *NewGUIDInteger(700))}
	for index := range expectedRanges {
		assert.True(t, expectedRanges[index].LowerGUID().Equals(*guidRanges[index].LowerGUID()),
			"Invalid lower GUID in range")
		assert.True(t, expectedRanges[index].HigherGUID().Equals(*guidRanges[index].HigherGUID()),
			"Invalid lower GUID in range")
	}

}

func TestRange_CreatePartitions_PartitionWithZero(t *testing.T) {
	lowerBound := int64(500)
	higherBound := int64(700)
	lowerGUID := NewGUIDInteger(lowerBound)
	higherGUID := NewGUIDInteger(higherBound)
	guidRange := NewGUIDRange(*lowerGUID, *higherGUID)
	partitions := []int{-1, 20, 20, 20, 20, 0, 10, 10, 0}

	guidRanges := guidRange.CreatePartitions(partitions)

	expectedRanges := []*Range{NewGUIDRange(*NewGUIDInteger(500), *NewGUIDInteger(540)),
		NewGUIDRange(*NewGUIDInteger(540), *NewGUIDInteger(580)),
		NewGUIDRange(*NewGUIDInteger(580), *NewGUIDInteger(620)),
		NewGUIDRange(*NewGUIDInteger(620), *NewGUIDInteger(660)),
		NewGUIDRange(*NewGUIDInteger(660), *NewGUIDInteger(680)),
		NewGUIDRange(*NewGUIDInteger(680), *NewGUIDInteger(700))}
	for index := range expectedRanges {
		assert.True(t, expectedRanges[index].LowerGUID().Equals(*guidRanges[index].LowerGUID()),
			"Invalid lower GUID in range")
		assert.True(t, expectedRanges[index].HigherGUID().Equals(*guidRanges[index].HigherGUID()),
			"Invalid lower GUID in range")
	}

}

func TestRange_Inside(t *testing.T) {
	lowerBound := int64(1500)
	higherBound := int64(1600)
	lowerGUID := NewGUIDInteger(lowerBound)
	higherGUID := NewGUIDInteger(higherBound)
	guidRange := NewGUIDRange(*lowerGUID, *higherGUID)

	insideIDs := []int64{1500, 1550, 1599}
	for _, insideID := range insideIDs {
		assert.True(t, guidRange.Inside(*NewGUIDInteger(insideID)), "GUID should be inside the range")
	}
	outsideIDs := []int64{1300, 1499, 1600, 1601, 1700}
	for _, outsideID := range outsideIDs {
		assert.False(t, guidRange.Inside(*NewGUIDInteger(outsideID)), "GUID should be outside the range")
	}
}

func TestRange_String(t *testing.T) {
	lowerBound := int64(1500)
	higherBound := int64(1600)
	lowerGUID := NewGUIDInteger(lowerBound)
	higherGUID := NewGUIDInteger(higherBound)
	guidRange := NewGUIDRange(*lowerGUID, *higherGUID)

	assert.Equal(t, fmt.Sprintf("[%d, %d)", lowerBound, higherBound), guidRange.String(),
		"Stringification of the range is wrong")
}
