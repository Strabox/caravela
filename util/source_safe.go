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

func (source *SourceSafe) Int63() int64 {
	source.mutex.Lock()
	defer source.mutex.Unlock()
	return source.source.Int63()
}

func (source *SourceSafe) Seed(seed int64) {
	source.mutex.Lock()
	defer source.mutex.Unlock()
	source.source.Seed(seed)
}
