package core

import (
	"context"
	"fmt"
	"time"
)

type ErrTooManyRequests struct {
	RetryAfter time.Duration
}

func (e ErrTooManyRequests) Error() string {
	return fmt.Sprintf("too many requests, retry after %s", e.RetryAfter)
}

type Algorithm interface {
	Reserve(tokens float64) error
	SortValue() string
	ExpireAt() time.Time
}

type AlgorithmStorer[T Algorithm] interface {
	Store(ctx context.Context, key string, alg T) (T, error)
	Load(ctx context.Context, key string) (*T, error)
}

type RateLimiter[alg Algorithm] struct {
	algStorer AlgorithmStorer[alg]
	new       func() alg
}

func NewRateLimiter[alg Algorithm](
	new func() alg,
	algStorer AlgorithmStorer[alg],
) RateLimiter[alg] {
	return RateLimiter[alg]{
		new:       new,
		algStorer: algStorer,
	}
}

func (r RateLimiter[Alg]) loadAlgorithm(ctx context.Context, key string) (Alg, error) {
	var defaultAlg Alg

	algorithm, err := r.algStorer.Load(ctx, key)
	if err != nil {
		return defaultAlg, fmt.Errorf("can't load data from key: %w", err)
	}

	if algorithm == nil {
		return r.new(), nil
	}

	return *algorithm, nil
}

func (r RateLimiter[Alg]) Reserve(ctx context.Context, key string, tokens float64) error {
	algorithm, err := r.loadAlgorithm(ctx, key)
	if err != nil {
		return err
	}

	err = algorithm.Reserve(tokens)
	if err != nil {
		return fmt.Errorf("can't reserve that capacity: %w", err)
	}

	_, err = r.algStorer.Store(ctx, key, algorithm)
	if err != nil {
		return fmt.Errorf("can't store key status: %w", err)
	}

	return nil
}
