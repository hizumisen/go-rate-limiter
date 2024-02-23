package benchmark

import (
	"context"
	"sync/atomic"

	"github.com/hizumisen/go-rate-limiter/core"
)

type storer[T core.Algorithm] struct {
	storeCount    atomic.Int32
	loadCount     atomic.Int32
	internalStore core.AlgorithmStorer[T]
}

func newMonitoredStorer[T core.Algorithm](internalStore core.AlgorithmStorer[T]) *storer[T] {
	return &storer[T]{
		internalStore: internalStore,
	}
}

func (store *storer[T]) Store(ctx context.Context, key string, alg T) (T, error) {
	store.storeCount.Add(1)
	return store.internalStore.Store(ctx, key, alg)
}

func (store *storer[T]) Load(ctx context.Context, key string) (*T, error) {
	store.loadCount.Add(1)
	return store.internalStore.Load(ctx, key)
}

func (store *storer[T]) fetchAndReset() (int32, int32) {
	storeCount := store.storeCount.Swap(0)
	loadCount := store.loadCount.Swap(0)
	return storeCount, loadCount
}
