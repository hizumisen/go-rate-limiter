package core

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type InMmemoryStore[T Algorithm] struct {
	data    map[string]T
	lock    sync.RWMutex
	maxSize int
}

var _ AlgorithmStorer[*TokenBucket] = &InMmemoryStore[*TokenBucket]{}

var ErrMaxSizeReached = errors.New("in memory max size reached")

func NewInMemoryStore[T Algorithm](maxSize int) *InMmemoryStore[T] {
	return &InMmemoryStore[T]{
		data:    make(map[string]T),
		lock:    sync.RWMutex{},
		maxSize: maxSize,
	}
}

func (m *InMmemoryStore[T]) Load(_ context.Context, key string) (*T, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	alg, ok := m.data[key]
	if !ok {
		return nil, nil
	}

	return &alg, nil
}

func (m *InMmemoryStore[T]) Store(_ context.Context, key string, alg T) (T, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	cached, ok := m.data[key]
	if !ok && len(m.data) >= m.maxSize {
		return alg, ErrMaxSizeReached
	}

	if ok &&
		cached.SortValue() > alg.SortValue() &&
		cached.ExpireAt().Before(time.Now()) {
		return alg, nil
	}

	m.data[key] = alg

	return alg, nil
}

func (m *InMmemoryStore[T]) Print() {
	fmt.Printf("[")

	for k, v := range m.data {
		fmt.Printf("\n\t%s - %v", k, v)
	}

	fmt.Printf("\n]\n")
}
