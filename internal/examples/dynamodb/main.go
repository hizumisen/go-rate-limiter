package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/hizumisen/go-rate-limiter/core"
	rateDynamodb "github.com/hizumisen/go-rate-limiter/dynamodb"
)

func main() {
	var store core.AlgorithmStorer[*core.TokenBucket]
	var client *dynamodb.Client
	var tableName string

	store = rateDynamodb.NewDynamoDbStore[*core.TokenBucket](client, tableName)

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
