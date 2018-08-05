package util

import (
	"math/rand"
	"sync"
)

// RandSafe is a version of math/rand that can be used by multiple goroutines safely.
type RandSafe struct {
	rand  *rand.Rand
	mutex sync.Mutex
}

// NewRandSafe creates a new RandSafe to be used.
func NewRandSafe(source rand.Source) *RandSafe {
	return &RandSafe{
		rand:  rand.New(source),
		mutex: sync.Mutex{},
	}
}

// =============================== math/rand rand methods =================================

func (r *RandSafe) ExpFloat64() float64 {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.rand.ExpFloat64()
}

func (r *RandSafe) Float32() float32 {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.rand.Float32()
}

func (r *RandSafe) Float64() float64 {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.rand.Float64()
}

func (r *RandSafe) Int() int {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.rand.Int()
}

func (r *RandSafe) Intn(n int) int {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.rand.Intn(n)
}

func (r *RandSafe) Int31() int32 {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.rand.Int31()
}

func (r *RandSafe) Int31n(n int32) int32 {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.rand.Int31n(n)
}

func (r *RandSafe) Int63() int64 {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.rand.Int63()
}

func (r *RandSafe) Int63n(n int64) int64 {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.rand.Int63n(n)
}

func (r *RandSafe) NormFloat64() float64 {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.rand.NormFloat64()
}

func (r *RandSafe) Perm(n int) []int {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.rand.Perm(n)
}

func (r *RandSafe) Read(p []byte) (int, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.rand.Read(p)
}

func (r *RandSafe) Seed(seed int64) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.rand.Seed(seed)
}

func (r *RandSafe) Uint32() uint32 {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.rand.Uint32()
}

func (r *RandSafe) Uint64() uint64 {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.rand.Uint64()
}
