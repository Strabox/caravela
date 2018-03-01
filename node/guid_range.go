package node

import (
	"fmt"
)

// GuidRange represents a guid range like [startId, endId)
type GuidRange struct {
	startId *Guid //Included in range
	endId   *Guid //Excluded from the range
}

func NewGuidRange(guid1 Guid, guid2 Guid) *GuidRange {
	return &GuidRange{&guid1, &guid2}
}

func (gr *GuidRange) GenerateRandomBetween() (*Guid, error) {
	return gr.startId.GenerateRandomBetween(*gr.endId)
}

func (gr *GuidRange) CreatePartitions(partitionsPerc []int) []*GuidRange {
	res := make([]*GuidRange, cap(partitionsPerc))

	currentBase := gr.startId.Copy()
	for index, percentage := range partitionsPerc {
		nextBase := currentBase.Copy()
		nextBase.AddOffset(gr.startId.Partitionate(percentage, *gr.endId))
		res[index] = NewGuidRange(*currentBase, *nextBase)
		currentBase = nextBase.Copy()
	}

	res[cap(partitionsPerc)-1].endId = gr.endId.Copy()
	return res
}

func (gr *GuidRange) Inside(guid Guid) bool {
	if (gr.startId.Cmp(guid) <= 0) && (gr.endId.Cmp(guid) >= 0) {
		return true
	}
	return false
}

func (gr *GuidRange) Print() {
	fmt.Printf("[%s, %s)", gr.startId.ToString(), gr.endId.ToString())
}
