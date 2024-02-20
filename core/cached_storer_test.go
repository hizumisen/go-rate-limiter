package core

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/hizumisen/go-rate-limiter/internal/testutils"
)

type storedItem time.Time

func newTimeAt(hour int) time.Time {
	return time.Date(2000, 1, 1, hour, 0, 0, 0, time.UTC)
}

func (s storedItem) Reserve(tokens float64) error {
	return nil
}

func (s storedItem) SortValue() string {
	return fmt.Sprintf("%v", s)
}

func (s storedItem) ExpireAt() time.Time {
	return time.Time(s)
}

type benchStorer[T any] struct {
	storeCount int
	loadCount  int
	alg        map[string]T
}

func newBenchStorer() *benchStorer[storedItem] {
	return &benchStorer[storedItem]{
		storeCount: 0,
		loadCount:  0,
		alg:        make(map[string]storedItem),
	}
}

func (store *benchStorer[T]) Store(ctx context.Context, key string, alg T) (T, error) {
	store.storeCount++
	return store.alg[key], nil
}

func (store *benchStorer[T]) Load(ctx context.Context, key string) (*T, error) {
	store.loadCount++
	alg := store.alg[key]
	return &alg, nil
}

func getKeys[T any](m map[string]T) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}

func TestCachedStore_Store_StoreIfNotCached(t *testing.T) {
	ctx := context.Background()
	logger := testutils.NewNoOpLogger()
	internalStore := newBenchStorer()
	store := NewCachedStore(
		ctx, logger, internalStore,
		10, 1*time.Second,
	)

	alg1 := storedItem(newTimeAt(1))
	internalStore.alg = map[string]storedItem{"key1": alg1}
	store.nowProvider = testutils.NowProvider(time.Date(1000, 1, 1, 0, 0, 0, 0, time.UTC))
	alg, err := store.Store(ctx, "key1", storedItem(newTimeAt(0)))
	testutils.RequireNoError(t, err)
	testutils.RequireEqual(t, alg1, alg)
	testutils.RequireEqual(t, 1, internalStore.storeCount)
}

func TestCachedStore_Store_NotStoreIfCached(t *testing.T) {
	ctx := context.Background()
	logger := testutils.NewNoOpLogger()
	internalStore := newBenchStorer()
	store := NewCachedStore(
		ctx, logger, internalStore,
		10, 1*time.Second,
	)

	internalStore.alg = map[string]storedItem{
		"key1": storedItem(newTimeAt(1)),
	}

	store.nowProvider = testutils.NowProvider(time.Date(1000, 1, 1, 0, 0, 0, 0, time.UTC))
	alg, err := store.Store(ctx, "key1", storedItem(newTimeAt(0)))
	testutils.RequireNoError(t, err)
	testutils.RequireEqual(t, storedItem(newTimeAt(1)), alg) //stored version
	testutils.RequireEqual(t, 1, internalStore.storeCount)

	store.nowProvider = testutils.NowProvider(time.Date(1001, 1, 1, 0, 0, 0, 0, time.UTC))
	alg, err = store.Store(ctx, "key1", storedItem(newTimeAt(2)))
	testutils.RequireNoError(t, err)
	testutils.RequireEqual(t, storedItem(newTimeAt(2)), alg) //cached version
	testutils.RequireEqual(t, 1, internalStore.storeCount)
}

func TestCachedStore_Store_FreeSpaceInTheCacheIfFull_RemoveLast(t *testing.T) {
	ctx := context.Background()
	logger := testutils.NewNoOpLogger()
	internalStore := newBenchStorer()
	store := NewCachedStore(
		ctx, logger, internalStore,
		2, 1*time.Hour,
	)

	internalStore.alg = map[string]storedItem{
		"key1": storedItem(newTimeAt(1)),
		"key2": storedItem(newTimeAt(2)),
		"key3": storedItem(newTimeAt(3)),
	}

	store.nowProvider = testutils.NowProvider(time.Date(1000, 1, 1, 0, 0, 0, 0, time.UTC))
	alg, err := store.Store(ctx, "key1", storedItem(newTimeAt(0)))
	testutils.RequireNoError(t, err)
	testutils.RequireEqual(t, storedItem(newTimeAt(1)), alg) //stored version
	testutils.RequireEqual(t, 1, internalStore.storeCount)
	testutils.RequireElementsMatch(t, getKeys(store.cache), []string{"key1"})

	store.nowProvider = testutils.NowProvider(time.Date(1000, 1, 1, 0, 0, 0, 1, time.UTC))
	alg, err = store.Store(ctx, "key2", storedItem(newTimeAt(0)))
	testutils.RequireNoError(t, err)
	testutils.RequireEqual(t, storedItem(newTimeAt(2)), alg) //stored version
	testutils.RequireEqual(t, 2, internalStore.storeCount)
	testutils.RequireElementsMatch(t, getKeys(store.cache), []string{"key1", "key2"})

	store.nowProvider = testutils.NowProvider(time.Date(1000, 1, 1, 0, 0, 0, 2, time.UTC))
	alg, err = store.Store(ctx, "key3", storedItem(newTimeAt(0)))
	testutils.RequireNoError(t, err)
	testutils.RequireEqual(t, storedItem(newTimeAt(3)), alg) //stored version
	testutils.RequireEqual(t, 3, internalStore.storeCount)
	testutils.RequireElementsMatch(t, getKeys(store.cache), []string{"key2", "key3"})
}

func TestCachedStore_Store_FreeSpaceInTheCacheIfFull_RemoveExpired(t *testing.T) {
	ctx := context.Background()
	logger := testutils.NewNoOpLogger()
	internalStore := newBenchStorer()
	store := NewCachedStore(
		ctx, logger, internalStore,
		2, 1*time.Hour,
	)

	internalStore.alg = map[string]storedItem{
		"key1": storedItem(newTimeAt(1)),
		"key2": storedItem(newTimeAt(2)),
		"key3": storedItem(newTimeAt(3)),
	}

	store.nowProvider = testutils.NowProvider(time.Date(1000, 1, 1, 0, 0, 0, 0, time.UTC))
	alg, err := store.Store(ctx, "key1", storedItem(newTimeAt(0)))
	testutils.RequireNoError(t, err)
	testutils.RequireEqual(t, storedItem(newTimeAt(1)), alg) //stored version
	testutils.RequireEqual(t, 1, internalStore.storeCount)
	testutils.RequireElementsMatch(t, getKeys(store.cache), []string{"key1"})

	alg, err = store.Store(ctx, "key2", storedItem(newTimeAt(0)))
	testutils.RequireNoError(t, err)
	testutils.RequireEqual(t, storedItem(newTimeAt(2)), alg) //stored version
	testutils.RequireEqual(t, 2, internalStore.storeCount)
	testutils.RequireElementsMatch(t, getKeys(store.cache), []string{"key1", "key2"})

	store.nowProvider = testutils.NowProvider(time.Date(1000, 1, 1, 1, 0, 0, 1, time.UTC))
	alg, err = store.Store(ctx, "key3", storedItem(newTimeAt(0)))
	testutils.RequireNoError(t, err)
	testutils.RequireEqual(t, storedItem(newTimeAt(3)), alg) //stored version
	testutils.RequireEqual(t, 3, internalStore.storeCount)
	testutils.RequireElementsMatch(t, getKeys(store.cache), []string{"key3"})
}

func TestCachedStore_Load_NotLoadIfCachedAndNotExpired(t *testing.T) {
	ctx := context.Background()
	logger := testutils.NewNoOpLogger()
	internalStore := newBenchStorer()
	store := NewCachedStore(
		ctx, logger, internalStore, 10,
		10*time.Hour, // long cache
	)

	internalStore.alg = map[string]storedItem{
		"key1": storedItem(newTimeAt(1)),
	}

	store.nowProvider = testutils.NowProvider(newTimeAt(1).Add(-10 * time.Hour)) //before expire
	alg1, err := store.Store(ctx, "key1", storedItem(newTimeAt(0)))
	testutils.RequireNoError(t, err)
	testutils.RequireEqual(t, storedItem(newTimeAt(1)), alg1)
	testutils.RequireEqual(t, 1, internalStore.storeCount)

	alg2, err := store.Load(ctx, "key1")
	testutils.RequireNoError(t, err)
	testutils.RequireEqual(t, storedItem(newTimeAt(1)), *alg2) //cached version
	testutils.RequireEqual(t, 1, internalStore.storeCount)
	testutils.RequireEqual(t, 0, internalStore.loadCount)
}

func TestCachedStore_Load_LoadIfCachedButNotUsed(t *testing.T) {
	ctx := context.Background()
	logger := testutils.NewNoOpLogger()
	internalStore := newBenchStorer()
	store := NewCachedStore(
		ctx, logger, internalStore, 10,
		1*time.Hour, //short cache
	)

	internalStore.alg = map[string]storedItem{
		"key1": storedItem(newTimeAt(1)),
	}

	store.nowProvider = testutils.NowProvider(newTimeAt(1).Add(-10 * time.Hour)) //before expire

	alg1, err := store.Store(ctx, "key1", storedItem(newTimeAt(0)))
	testutils.RequireNoError(t, err)
	testutils.RequireEqual(t, storedItem(newTimeAt(1)), alg1)
	testutils.RequireEqual(t, 1, internalStore.storeCount)

	internalStore.alg = map[string]storedItem{
		"key1": storedItem(newTimeAt(2)),
	}

	store.nowProvider = testutils.NowProvider(store.nowProvider().Add(2 * time.Hour)) //before expire but after cache duration
	alg2, err := store.Load(ctx, "key1")
	testutils.RequireNoError(t, err)
	testutils.RequireEqual(t, storedItem(newTimeAt(2)), *alg2) //store version
	testutils.RequireEqual(t, 1, internalStore.loadCount)
}

func TestCachedStore_Load_LoadIfCachedButExpired(t *testing.T) {
	ctx := context.Background()
	logger := testutils.NewNoOpLogger()
	internalStore := newBenchStorer()
	store := NewCachedStore(
		ctx, logger, internalStore,
		10,
		10*time.Hour, //long cache
	)

	internalStore.alg = map[string]storedItem{
		"key1": storedItem(newTimeAt(1)),
	}

	store.nowProvider = testutils.NowProvider(newTimeAt(1).Add(-1 * time.Hour)) //before expire
	alg1, err := store.Store(ctx, "key1", storedItem(newTimeAt(0)))
	testutils.RequireNoError(t, err)
	testutils.RequireEqual(t, storedItem(newTimeAt(1)), alg1)
	testutils.RequireEqual(t, 1, internalStore.storeCount)

	internalStore.alg = map[string]storedItem{
		"key1": storedItem(newTimeAt(2)),
	}

	store.nowProvider = testutils.NowProvider(newTimeAt(1).Add(1 * time.Hour)) //after expire
	alg2, err := store.Load(ctx, "key1")
	testutils.RequireNoError(t, err)
	testutils.RequireEqual(t, storedItem(newTimeAt(2)), *alg2) //store version
	testutils.RequireEqual(t, 1, internalStore.loadCount)
}

func TestCachedStore_Load_LoadIfNotCached(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	internalStore := newBenchStorer()
	store := NewCachedStore(
		ctx, logger, internalStore,
		10, 1*time.Second,
	)

	internalStore.alg = map[string]storedItem{
		"key1": storedItem(newTimeAt(1)),
	}

	store.nowProvider = testutils.NowProvider(time.Date(1001, 1, 1, 0, 0, 0, 0, time.UTC))
	alg2, err := store.Load(ctx, "key1")
	testutils.RequireNoError(t, err)
	testutils.RequireEqual(t, storedItem(newTimeAt(1)), *alg2) //store version
	testutils.RequireEqual(t, 1, internalStore.loadCount)
}
