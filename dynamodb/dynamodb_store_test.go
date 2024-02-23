package dynamodb_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hizumisen/go-rate-limiter/dynamodb"
	"github.com/hizumisen/go-rate-limiter/internal/testutils"
)

type storedItem time.Time

func newStoredItemAtHour(hour int) storedItem {
	return storedItem(time.Date(3000, 1, 1, hour, 0, 0, 0, time.UTC))
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

func (s storedItem) String() string {
	return time.Time(s).Format(time.Layout)
}

func buildStore(ctx context.Context, t *testing.T) *dynamodb.DynamoDbStore[storedItem] {
	t.Helper()

	dyanamodbClient, err := dynamodb.NewDynamodbClient(ctx)
	testutils.RequireNoError(t, err)
	tableName := fmt.Sprintf("rate-limit-%d", time.Now().UnixNano())
	dynamodb.CreateTableIfMissing(ctx, dyanamodbClient, tableName, dynamodb.GetTableConfiguration())
	return dynamodb.NewDynamoDbStore[storedItem](dyanamodbClient, tableName)
}

func TestDynamoDbStore_Store_NewAlg(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := buildStore(ctx, t)

	got, err := store.Store(ctx, "key1", newStoredItemAtHour(1))
	testutils.RequireNoError(t, err)
	testutils.RequireEqual(t, newStoredItemAtHour(1), got)
}

func TestDynamoDbStore_Store_OverrideAlgIfSortIsGreater(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := buildStore(ctx, t)

	got, err := store.Store(ctx, "key1", newStoredItemAtHour(1))
	testutils.RequireNoError(t, err)
	testutils.RequireEqual(t, newStoredItemAtHour(1), got)

	got, err = store.Store(ctx, "key1", newStoredItemAtHour(2))
	testutils.RequireNoError(t, err)
	testutils.RequireEqual(t, newStoredItemAtHour(2), got)
}

func TestDynamoDbStore_Store_NotOverrideAlgIfSortIsLesserAndReturnStored(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := buildStore(ctx, t)

	got, err := store.Store(ctx, "key1", newStoredItemAtHour(2))
	testutils.RequireNoError(t, err)
	testutils.NewNoOpLogger()
	testutils.RequireEqual(t, newStoredItemAtHour(2), got)

	got, err = store.Store(ctx, "key1", newStoredItemAtHour(1))
	testutils.RequireNoError(t, err)
	testutils.RequireEqual(t, newStoredItemAtHour(2), got)
}
