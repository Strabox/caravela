package partitions

import (
	"math"
	"math/rand"
	"sync"
)

type ResourcePartitionState struct {
	totalTries      int
	hits            int
	mutex           sync.RWMutex
	randomGenerator *rand.Rand
}

func NewResourcePartitionState(totalStat int, randGenerator *rand.Rand) *ResourcePartitionState {
	return &ResourcePartitionState{
		totalTries:      totalStat,
		hits:            totalStat,
		mutex:           sync.RWMutex{},
		randomGenerator: randGenerator,
	}
}

func (r *ResourcePartitionState) Try() bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	hitProbability := int((float64(r.hits) / float64(r.totalTries)) * 100)
	randTry := r.randomGenerator.Intn(100)
	if randTry <= hitProbability {
		return true
	}
	lastChance := r.randomGenerator.Intn(100)
	if lastChance <= 9 {
		return true
	}
	return false
}

func (r *ResourcePartitionState) Hit() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.hits < r.totalTries {
		r.hits++
	}
}

func (r *ResourcePartitionState) Miss() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.hits > 0 {
		r.hits--
	}
}

func (r *ResourcePartitionState) Merge(newHits int) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.hits = int(math.Floor((float64(newHits) + float64(r.hits)) / 2))
}
