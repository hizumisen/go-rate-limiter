package benchmark

import (
	"github.com/hizumisen/go-rate-limiter/core"
)

func Run() {
	configs := []tokenBucketRunConfig{
		{
			store:            dynamodbStore[*core.TokenBucket](),
			receiveBurst:     10.0,
			receivePerSecond: 10.0,
			receivers:        4,
			keys:             1,
		},
		{
			store:            dynamodbStore[*core.TokenBucket](),
			receiveBurst:     10.0,
			receivePerSecond: 10.0,
			receivers:        4,
			keys:             10,
		},
	}

	for _, config := range configs {
		runTokenBucket(config)
	}

}
