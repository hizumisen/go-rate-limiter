# go-rate-limiter
Distributed rate limiter written go with minimal dependencies.

# Overview
This project provides a simple yet efficient distributed rate limiter for Go applications. It consists of two modules:

**core**: It includes essential components for rate limiting and an in-memory rate limiter implementation using only the Go standard library.

**dynamodb**: An extension module that provides a DynamoDB-backed implementation of the rate limiter for distributed environments.

# Core module

### Installation
```go
go get github.com/hizumisen/go-rate-limiter/core
```

### Usage
To use the rate limiter
```go
import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hizumisen/go-rate-limiter/core"
)

func main() {
	ctx := context.Background()
	
	// Create an AgorithmStorer based on the implementation chosen
	var store core.AlgorithmStorer[*core.TokenBucket]
	
	requestBurst := 100.0
	requestPerSecond := 10.0
	
	rateLimit := core.NewRateLimiter(
		func() *core.TokenBucket {
			return core.NewTokenBucket(requestBurst, requestPerSecond)
		},
		store,
	)

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
```

To create the in memory storer
```go
import "github.com/hizumisen/go-rate-limiter/core"

storeSize := 100
store := core.NewInMemoryStore[*core.TokenBucket](storeSize)
```

# DynamoDB module

### Installation
```go
go get github.com/hizumisen/go-rate-limiter/core
go get github.com/hizumisen/go-rate-limiter/dynamodb
```

### Usage
To create the DynamoDB storer

```go
import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/hizumisen/go-rate-limiter/core"
	rateDynamodb "github.com/hizumisen/go-rate-limiter/dynamodb"
)

// Create a DynamoDB client
var client *dynamodb.Client
var tableName string

store := rateDynamodb.NewDynamoDbStore[*core.TokenBucket](client, tableName)
```