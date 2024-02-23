package benchmark

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hizumisen/go-rate-limiter/core"
)

type tokenBucketRunConfig struct {
	name             string
	store            core.AlgorithmStorer[*core.TokenBucket]
	requestPerSecond float64
	receiveBurst     float64
	receivePerSecond float64
	receivers        int
	keys             int
}

func runTokenBucket(config tokenBucketRunConfig) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	requestPerSecond := float64(config.keys) * (config.receiveBurst + config.receivePerSecond) * 1.1
	fmt.Printf("request per second %f\n", requestPerSecond)

	metrics, err := run(
		ctx,
		config.store,
		func() *core.TokenBucket {
			return core.NewTokenBucket(config.receiveBurst, config.receivePerSecond)
		},
		requestPerSecond,
		config.receivers,
		config.keys,
	)
	if err != nil {
		log.Fatal(err)
	}

	render(metrics, config)
}
