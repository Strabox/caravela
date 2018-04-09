package guid

import (
	log "github.com/Sirupsen/logrus"
)

// Range represents a guid range like [startId, endId)
type Range struct {
	startId *Guid //Included in range
	endId   *Guid //Excluded from the range
}

func NewGuidRange(guid1 Guid, guid2 Guid) *Range {
	return &Range{&guid1, &guid2}
}

func (gr *Range) GenerateRandomBetween() (*Guid, error) {
	return gr.startId.GenerateRandomBetween(*gr.endId)
}

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

func (gr *Range) Inside(guid Guid) bool {
	if (gr.startId.Cmp(guid) <= 0) && (gr.endId.Cmp(guid) > 0) {
		return true
	}
	return false
}

func (gr *Range) Print() {
	log.Printf("[%s, %s)", gr.startId.String(), gr.endId.String())
}
