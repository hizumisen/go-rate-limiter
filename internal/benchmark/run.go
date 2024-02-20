package benchmark

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/hizumisen/go-rate-limiter/core"

	"golang.org/x/sync/errgroup"
)

type RateLimiter interface {
	Reserve(ctx context.Context, key string, tokens float64) error
}

type result struct {
	startedAt  time.Time
	endedAt    time.Time
	accepted   bool
	retryAfter time.Duration
}

type metric struct {
	time     int
	accepted int32
	rejected int32
	stored   int32
	loaded   int32
}

func sender(
	ctx context.Context,
	wait time.Duration,
	requests chan<- time.Time,
) error {
	ticker := time.NewTicker(wait)

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context expired: %w", ctx.Err())
		case <-ticker.C:
			select {
			case <-ctx.Done():
				return fmt.Errorf("context expired: %w", ctx.Err())
			case requests <- time.Now():
			}
		}
	}
}

func receiver(
	ctx context.Context,
	requests <-chan time.Time,
	results chan<- result,
	rateLimiter RateLimiter,
	keys int,
) error {
	key := 0
	for startedAt := range requests {
		err := rateLimiter.Reserve(ctx, fmt.Sprintf("key_%d", key), 1)
		key = (key + 1) % keys
		var res *result

		var tooManyReqErr core.ErrTooManyRequests
		switch {
		case errors.As(err, &tooManyReqErr):
			res = &result{
				startedAt:  startedAt,
				endedAt:    time.Now(),
				accepted:   false,
				retryAfter: tooManyReqErr.RetryAfter,
			}
		case err != nil:
			return fmt.Errorf("unexpected error from reserve: %w", err)
		default:
			res = &result{
				startedAt: startedAt,
				endedAt:   time.Now(),
				accepted:  true,
			}
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("context expired: %w", ctx.Err())
		case results <- *res:
		default:
		}
	}

	return nil
}

func reporter[T core.Algorithm](
	ctx context.Context,
	monitoredStorer *storer[T],
	results <-chan result,
) ([]metric, error) {
	var acceptCounter atomic.Int32
	var rejectCounter atomic.Int32
	start := time.Now()
	lastBucket := start

	metrics := []metric{}

	for result := range results {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context expired: %w", ctx.Err())
		default:
		}

		if result.accepted {
			acceptCounter.Add(1)
		} else {
			rejectCounter.Add(1)
		}

		if time.Since(lastBucket) >= 1*time.Second {
			lastBucket = time.Now()
			storeCount, loadCount := monitoredStorer.fetchAndReset()
			acceptCount := acceptCounter.Swap(0)
			rejectCount := rejectCounter.Swap(0)
			currentMetric := metric{
				time:     int(time.Since(start).Seconds()),
				accepted: acceptCount,
				rejected: rejectCount,
				stored:   storeCount,
				loaded:   loadCount,
			}

			metrics = append(metrics, currentMetric)
		}

	}

	return metrics, nil
}

func run[alg core.Algorithm](
	ctx context.Context,
	store core.AlgorithmStorer[alg],
	new func() alg,
	requestPerSecond float64,
	receivers int,
	keys int,
) ([]metric, error) {
	monitoredStorer := newMonitoredStorer(store)
	g, ctx := errgroup.WithContext(ctx)
	requests := make(chan time.Time, 10000)
	results := make(chan result, 10000)
	metrics := make(chan []metric, 1)
	waitBetweenRequest := time.Duration(1000000000 / requestPerSecond)

	receiverG, _ := errgroup.WithContext(ctx)
	for i := 0; i < receivers; i++ {
		receiverG.Go(func() error {
			rateLimiter := core.NewRateLimiter(new, monitoredStorer)
			return receiver(ctx, requests, results, rateLimiter, keys)
		})
	}

	g.Go(func() error {
		defer close(results)
		return receiverG.Wait()
	})

	g.Go(func() error {
		defer close(metrics)
		m, err := reporter(ctx, monitoredStorer, results)
		metrics <- m
		return err
	})

	//start the sender once the rest of the go routines are up
	g.Go(func() error {
		defer close(requests)
		return sender(ctx, waitBetweenRequest, requests)
	})

	if err := g.Wait(); err != nil {
		if !errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
	}

	metricsArray := <-metrics
	fmt.Println("test finished")
	for _, metric := range metricsArray {
		fmt.Printf(
			"%v accepted=%v rejected=%v stored=%v loaded=%v\n",
			metric.time,
			metric.accepted,
			metric.rejected,
			metric.loaded,
			metric.stored,
		)
	}

	return metricsArray, nil
}
