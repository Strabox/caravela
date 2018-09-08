package guid

import (
	"fmt"
)

// Range represents a range of GUIDs (Global Unique Identifiers), i.e. [lowerGUID, higherGUID).
type Range struct {
	lowerGUID  *GUID // Lower GUID of the range, included in the range.
	higherGUID *GUID // Higher GUID of the range, excluded from the range.
}

// NewGUIDRange creates a new GUID range given a lower GUID and higher GUID.
func NewGUIDRange(lowerGUID GUID, higherGUID GUID) *Range {
	return &Range{
		lowerGUID:  &lowerGUID,
		higherGUID: &higherGUID,
	}
}

// GenerateRandom generate random GUID inside the range.
func (r *Range) GenerateRandom() (*GUID, error) {
	return r.lowerGUID.GenerateInnerRandomGUIDScaled(*r.higherGUID)
}

// CreatePartitions returns partitions, set of ranges, of the receiver range.
func (r *Range) CreatePartitions(partitionsPercentage []int) []*Range {
	res := make([]*Range, 0)

	currentBase := r.lowerGUID.Copy()
	for _, percentage := range partitionsPercentage {
		if percentage <= 0 || percentage > 100 {
			continue
		}
		nextBase := currentBase.Copy()
		nextBase.AddOffset(r.lowerGUID.PercentageOffset(percentage, *r.higherGUID))
		res = append(res, NewGUIDRange(*currentBase, *nextBase))
		currentBase = nextBase.Copy()
	}

	res[len(res)-1].higherGUID = r.higherGUID.Copy()
	return res
}

// Inside verify if the given GUID is inside the range.
func (r *Range) Inside(guid GUID) bool {
	if (r.lowerGUID.Cmp(guid) <= 0) && (r.higherGUID.Cmp(guid) > 0) {
		return true
	}
	return false
}

// LowerGUID get the lower GUID of the range.
func (r *Range) LowerGUID() *GUID {
	return r.lowerGUID.Copy()
}

// HigherGUID get the higher GUID of the range.
func (r *Range) HigherGUID() *GUID {
	return r.higherGUID.Copy()
}

// String returns the string representation of the range.
func (r *Range) String() string {
	return fmt.Sprintf("[%s, %s)", r.lowerGUID.String(), r.higherGUID.String())
}
