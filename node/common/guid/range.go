package guid

import log "github.com/Sirupsen/logrus"

/*
Range represents a range of GUIDs i.e. [startId, endId)
*/
type Range struct {
	startId *Guid //Included in range
	endId   *Guid //Excluded from the range
}

/*
Creates a new GUID range given a start GUID and end GUID
*/
func NewGuidRange(startGUID Guid, endGUID Guid) *Range {
	return &Range{&startGUID, &endGUID}
}

/*
Generate random GUID inside the range.
*/
func (gr *Range) GenerateRandomBetween() (*Guid, error) {
	return gr.startId.GenerateRandomBetween(*gr.endId)
}

/*
Create partitions of the GUID range.
*/
func (gr *Range) CreatePartitions(partitionsPerc []int) []*Range {
	res := make([]*Range, cap(partitionsPerc))

	currentBase := gr.startId.Copy()
	for index, percentage := range partitionsPerc {
		nextBase := currentBase.Copy()
		nextBase.AddOffset(gr.startId.Partitioning(percentage, *gr.endId))
		res[index] = NewGuidRange(*currentBase, *nextBase)
		currentBase = nextBase.Copy()
	}

	res[cap(partitionsPerc)-1].endId = gr.endId.Copy()
	return res
}

/*
Verify if the given GUID is inside the range.
*/
func (gr *Range) Inside(guid Guid) bool {
	if (gr.startId.Cmp(guid) <= 0) && (gr.endId.Cmp(guid) > 0) {
		return true
	}
	return false
}

/*
Get the start GUID of the range.
*/
func (gr *Range) StartGUID() *Guid {
	return gr.startId.Copy()
}

/*
Get the end GUID of the range.
*/
func (gr *Range) EndGUID() *Guid {
	return gr.endId.Copy()
}

/*
Print the range into the log.
*/
func (gr *Range) Print() {
	log.Debugf("[%s, %s)", gr.startId.String(), gr.endId.String())
}
