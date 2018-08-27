package util

import (
	"math/rand"
	"sync"
)

// SourceSafe is a version of math/rand source that can be used by multiple goroutines safely.
type SourceSafe struct {
	source rand.Source
	mutex  sync.Mutex
}

// NewSourceSafe creates a new SourceSafe to be used.
func NewSourceSafe(source rand.Source) *SourceSafe {
	return &SourceSafe{
		source: source,
		mutex:  sync.Mutex{},
	}
}

// =============================== math/rand source methods =================================

func (s *SourceSafe) Int63() int64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.source.Int63()
}

func (s *SourceSafe) Seed(seed int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.source.Seed(seed)
}
