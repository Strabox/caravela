package supplier

import (
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/util"
	"math"
	"math/rand"
	"sync"
	"time"
)

var randomGenerator = rand.New(util.NewSourceSafe(rand.NewSource(time.Now().Unix())))

type SystemResourcePartitions struct {
	partitionsState sync.Map
	totalStats      int
}

func NewSystemResourcePartitions(totalStats int) *SystemResourcePartitions {
	return &SystemResourcePartitions{
		partitionsState: sync.Map{},
		totalStats:      totalStats,
	}
}

func (s *SystemResourcePartitions) Try(targetResPartition resources.Resources) bool {
	if partition, exist := s.partitionsState.Load(targetResPartition); exist {
		if partitionState, ok := partition.(*ResourcePartitionState); ok {
			return partitionState.Try()
		}
	} else {
		newPartitionState := NewResourcePartitionState(s.totalStats)
		s.partitionsState.Store(targetResPartition, newPartitionState)
		return newPartitionState.Try()
	}
	return true
}

func (s *SystemResourcePartitions) Hit(resPartition resources.Resources) {
	if partition, exist := s.partitionsState.Load(resPartition); exist {
		if partitionState, ok := partition.(*ResourcePartitionState); ok {
			partitionState.Hit()
		}
	}
}

func (s *SystemResourcePartitions) Miss(resPartition resources.Resources) {
	if partition, exist := s.partitionsState.Load(resPartition); exist {
		if partitionState, ok := partition.(*ResourcePartitionState); ok {
			partitionState.Miss()
		}
	}
}

func (s *SystemResourcePartitions) PartitionsState() []types.PartitionState {
	res := make([]types.PartitionState, 0)
	s.partitionsState.Range(func(key, value interface{}) bool {
		partResources, _ := key.(resources.Resources)
		if partitionState, ok := value.(*ResourcePartitionState); ok {
			res = append(res, types.PartitionState{
				PartitionResources: types.Resources{
					CPUs: partResources.CPUs(),
					RAM:  partResources.RAM(),
				},
				Hits: partitionState.hits,
			})
		}
		return true
	})
	return res
}

func (s *SystemResourcePartitions) MergePartitionsState(newPartitionsState []types.PartitionState) {
	for _, newPartitionState := range newPartitionsState {
		partRes := resources.NewResources(newPartitionState.PartitionResources.CPUs, newPartitionState.PartitionResources.RAM)
		if partition, exist := s.partitionsState.Load(*partRes); exist {
			if partitionState, ok := partition.(*ResourcePartitionState); ok {
				partitionState.Merge(newPartitionState.Hits)
			}
		} else {
			unknownPartitionState := NewResourcePartitionState(s.totalStats)
			unknownPartitionState.hits = newPartitionState.Hits
			s.partitionsState.Store(*partRes, unknownPartitionState)
		}
	}
}

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
