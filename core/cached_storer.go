package core

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type cachedItem[T any] struct {
	alg        T
	sort       string
	lastUsedAt time.Time
}

func newCachedItem[T Algorithm](
	alg T,
	lastUsedAt time.Time,
) cachedItem[T] {
	return cachedItem[T]{
		alg:        alg,
		sort:       alg.SortValue(),
		lastUsedAt: lastUsedAt,
	}
}

type CachedStore[T Algorithm] struct {
	internalCtx   context.Context
	logger        *slog.Logger
	actualStore   AlgorithmStorer[T]
	cache         map[string]cachedItem[T]
	lock          sync.RWMutex
	cacheSize     int
	cacheDuration time.Duration
	cancel        func()
	nowProvider   func() time.Time
}

func NewCachedStore[T Algorithm](
	ctx context.Context,
	logger *slog.Logger,
	actualStore AlgorithmStorer[T],
	cacheSize int,
	cacheDuration time.Duration,
) *CachedStore[T] {
	store := &CachedStore[T]{
		internalCtx:   ctx,
		logger:        logger,
		actualStore:   actualStore,
		cache:         make(map[string]cachedItem[T]),
		lock:          sync.RWMutex{},
		cacheSize:     cacheSize,
		cacheDuration: cacheDuration,
		nowProvider:   time.Now,
	}

	store.start()

	return store
}

func (store *CachedStore[T]) start() {
	store.cancel = runPeriodically(store.cacheDuration, store.flushAndRefreshData)
}

func (store *CachedStore[T]) Stop() {
	store.flushAndRefreshData()
	if store != nil {
		store.cancel()
	}
}

func runPeriodically(duration time.Duration, fun func()) func() {
	done := make(chan bool)
	ticker := time.NewTicker(duration)

	go func() {
		for {
			select {
			case <-done:
				ticker.Stop()
				return
			case <-ticker.C:
				fun()
			}
		}
	}()

	cancelFun := func() {
		done <- true
	}

	return cancelFun
}

func (store *CachedStore[T]) isExpiredFromCache(item cachedItem[T]) bool {
	return store.nowProvider().After(item.alg.ExpireAt()) ||
		store.nowProvider().Sub(item.lastUsedAt) > store.cacheDuration
}

func (store *CachedStore[T]) flushAndRefreshData() {
	store.lock.Lock()
	defer store.lock.Unlock()

	var errs []error

	store.removeExpiredFromCache()

	for key, val := range store.cache {
		if val.sort != "" {
			updated, err := store.actualStore.Store(store.internalCtx, key, val.alg)
			if err != nil {
				errs = append(errs, err)
			}

			store.cache[key] = newCachedItem(updated, val.lastUsedAt)
		}
	}

	if len(errs) > 0 {
		store.logger.Warn("can't persist and refresh cached data", "errors", errs)
	}
}

func (store *CachedStore[T]) removeExpiredFromCache() bool {
	removedSomething := false

	for key, val := range store.cache {
		if store.isExpiredFromCache(val) {
			delete(store.cache, key)
			removedSomething = true
		}
	}

	return removedSomething
}

func (store *CachedStore[T]) removeLastCached() {
	var minKey string

	for key := range store.cache {
		minKey = key
		break
	}

	for key, alg := range store.cache {
		if store.cache[minKey].lastUsedAt.After(alg.lastUsedAt) {
			minKey = key
		}
	}

	delete(store.cache, minKey)
}

func (store *CachedStore[T]) Store(ctx context.Context, key string, alg T) (T, error) {
	store.lock.Lock()
	defer store.lock.Unlock()

	_, ok := store.cache[key]
	if ok {
		store.cache[key] = newCachedItem(alg, store.nowProvider())
		return alg, nil
	}

	//free some space in the cache
	if len(store.cache) >= store.cacheSize {
		removed := store.removeExpiredFromCache()
		if !removed {
			store.removeLastCached()
		}
	}

	//refresh from inner store
	alg, err := store.actualStore.Store(ctx, key, alg)
	if err != nil {
		return alg, fmt.Errorf("can't store alg: %w", err)
	}

	if store.cacheSize > 0 {
		store.cache[key] = newCachedItem(alg, store.nowProvider())
	}

	return alg, nil
}

func (store *CachedStore[T]) Load(ctx context.Context, key string) (*T, error) {
	store.lock.Lock()
	defer store.lock.Unlock()

	val, ok := store.cache[key]
	if ok && !store.isExpiredFromCache(val) {
		return &val.alg, nil
	}

	alg, err := store.actualStore.Load(ctx, key)
	if err != nil {
		return alg, fmt.Errorf("can't load alg: %w", err)
	}

	if alg == nil {
		return nil, nil
	}

	if store.cacheSize > 0 {
		store.cache[key] = newCachedItem(*alg, store.nowProvider())
	}

	return alg, nil
}
