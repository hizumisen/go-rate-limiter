package core_test

import (
	"testing"

	"github.com/hizumisen/go-rate-limiter/core"
	"github.com/hizumisen/go-rate-limiter/internal/testutils"
)

func TestTokenBucket_SortValue_NotDecreaseAfterReserve(t *testing.T) {
	token := core.NewTokenBucket(10, 1)
	sortValue1 := token.SortValue()
	err := token.Reserve(1)
	testutils.RequireNoError(t, err)
	sortValue2 := token.SortValue()

	if sortValue1 > sortValue2 {
		t.Errorf("TokenBucket.SortValue() = %v is less than %v", sortValue1, sortValue2)
	}
}

func TestTokenBucket_SortValue_CreaseMoreIfMoreIsReserved(t *testing.T) {
	time1 := testutils.NewTimeAt(1)
	token1 := core.NewTokenBucket(10, 1).WitNowProvider(testutils.NowProvider(time1))
	token2 := core.NewTokenBucket(10, 1).WitNowProvider(testutils.NowProvider(time1))

	err := token1.Reserve(1)
	testutils.RequireNoError(t, err)
	err = token2.Reserve(5)
	testutils.RequireNoError(t, err)

	sortValue1 := token1.SortValue()
	sortValue2 := token2.SortValue()

	if sortValue1 >= sortValue2 {
		t.Errorf("TokenBucket.SortValue() = %v is less or equal than %v", sortValue1, sortValue2)
	}
}
