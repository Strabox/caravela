package guid

import (
	"fmt"
)

/*
Represents a range of GUIDs i.e. [lowerGUID, higherGUID)
*/
type Range struct {
	lowerGUID  *GUID //Included in range
	higherGUID *GUID //Excluded from the range
}

/*
Creates a new GUID range given a lower GUID and higher GUID.
*/
func NewGUIDRange(lowerGUID GUID, higherGUID GUID) *Range {
	return &Range{lowerGUID: &lowerGUID, higherGUID: &higherGUID}
}

/*
Generate random GUID inside the range.
*/
func (gr *Range) GenerateRandomInside() (*GUID, error) {
	return gr.lowerGUID.GenerateInnerRandomGUID(*gr.higherGUID)
}

/*
Create partitions (set of ranges) of the receiver range.
*/
func (gr *Range) CreatePartitions(partitionsPercentage []int) []*Range {
	res := make([]*Range, 0)

	currentBase := gr.lowerGUID.Copy()
	for _, percentage := range partitionsPercentage {
		if percentage <= 0 || percentage > 100 {
			continue
		}
		nextBase := currentBase.Copy()
		nextBase.AddOffset(gr.lowerGUID.PercentageOffset(percentage, *gr.higherGUID))
		res = append(res, NewGUIDRange(*currentBase, *nextBase))
		currentBase = nextBase.Copy()
	}

	res[len(res)-1].higherGUID = gr.higherGUID.Copy()
	return res
}

/*
Verify if the given GUID is inside the range.
*/
func (gr *Range) Inside(guid GUID) bool {
	if (gr.lowerGUID.Cmp(guid) <= 0) && (gr.higherGUID.Cmp(guid) > 0) {
		return true
	}
	return false
}

/*
Get the lower GUID of the range.
*/
func (gr *Range) LowerGUID() *GUID {
	return gr.lowerGUID.Copy()
}

/*
Get the higher GUID of the range.
*/
func (gr *Range) HigherGUID() *GUID {
	return gr.higherGUID.Copy()
}

/*
Print the range into the log.
*/
func (gr *Range) String() string {
	return fmt.Sprintf("[%s, %s)", gr.lowerGUID.String(), gr.higherGUID.String())
}
