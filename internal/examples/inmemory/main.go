package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hizumisen/go-rate-limiter/core"
)

func main() {
	storeSize := 100
	store := core.NewInMemoryStore[*core.TokenBucket](storeSize)

	requestBurst := 100.0
	requestPerSecond := 10.0

	rateLimit := core.NewRateLimiter(
		func() *core.TokenBucket {
			return core.NewTokenBucket(requestBurst, requestPerSecond)
		},
		store,
	)

	ctx := context.Background()

	for {
		err := rateLimit.Reserve(ctx, "key", 1)
		if err != nil {
			var tooManyReqErr core.ErrTooManyRequests
			if errors.As(err, &tooManyReqErr) {
				log.Fatalf("request denied: retry after %s\n", tooManyReqErr.RetryAfter)
			}

			log.Fatal(err)
		}

		fmt.Println("request accepted")
	}
}
