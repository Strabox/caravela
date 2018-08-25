package partitions

import (
	"github.com/strabox/caravela/util"
	"math"
	"math/rand"
	"sync"
	"time"
)

var randomGenerator = rand.New(util.NewSourceSafe(rand.NewSource(time.Now().Unix())))

type ResourcePartitionState struct {
	totalTries int
	hits       int
	mutex      sync.RWMutex
}

func NewResourcePartitionState(totalStat int) *ResourcePartitionState {
	return &ResourcePartitionState{
		totalTries: totalStat,
		hits:       totalStat,
		mutex:      sync.RWMutex{},
	}
}

func (rps *ResourcePartitionState) Try() bool {
	rps.mutex.RLock()
	defer rps.mutex.RUnlock()

	hitProbability := int((float64(rps.hits) / float64(rps.totalTries)) * 100)
	randTry := randomGenerator.Intn(100)
	if randTry <= hitProbability {
		return true
	}
	lastChance := randomGenerator.Intn(100)
	if lastChance <= 10 {
		return true
	}
	return false
}

func (rps *ResourcePartitionState) Hit() {
	rps.mutex.Lock()
	defer rps.mutex.Unlock()

	if rps.hits < rps.totalTries {
		rps.hits++
	}
}

func (rps *ResourcePartitionState) Miss() {
	rps.mutex.Lock()
	defer rps.mutex.Unlock()

	if rps.hits > 0 {
		rps.hits--
	}
}

func (rps *ResourcePartitionState) Merge(newHits int) {
	rps.mutex.Lock()
	defer rps.mutex.Unlock()

	rps.hits = int(math.Floor((float64(newHits) + float64(rps.hits)) / 2))
}
