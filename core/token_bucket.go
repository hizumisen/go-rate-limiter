package core

import (
	"errors"
	"fmt"
	"math"
	"time"
)

type TokenBucket struct {
	Tokens         float64
	MaxTokens      float64
	RefillRate     float64
	LastRefillTime time.Time
	nowProvider    func() time.Time //for test
}

var _ Algorithm = &TokenBucket{}

var errOutOfBoundsRequest = errors.New("capacity requested is greater than the maximum allowed")

func NewTokenBucket(maxTokens, refillRate float64) *TokenBucket {
	return &TokenBucket{
		Tokens:         maxTokens,
		MaxTokens:      maxTokens,
		RefillRate:     refillRate,
		LastRefillTime: time.Now(),
	}
}

func (tb *TokenBucket) WitNowProvider(fun func() time.Time) *TokenBucket {
	tb.LastRefillTime = fun()
	tb.nowProvider = fun
	return tb
}

func (tb *TokenBucket) now() time.Time {
	if tb.nowProvider != nil {
		return tb.nowProvider()
	} else {
		return time.Now()
	}
}

func (tb *TokenBucket) refill() {
	now := tb.now()
	duration := now.Sub(tb.LastRefillTime)
	//how many tokens have been added in bucket since last refill call
	tokensToAdd := tb.RefillRate * duration.Seconds()
	tb.Tokens = math.Min(tb.Tokens+tokensToAdd, tb.MaxTokens)
	tb.LastRefillTime = now
}

func (tb *TokenBucket) howMuchToWaitFor(tokens float64) time.Duration {
	remainingTokens := math.Max(tokens-tb.Tokens, 0)
	durationSeconds := remainingTokens / tb.RefillRate
	return time.Duration(durationSeconds * 1000000000)
}

func (tb *TokenBucket) Reserve(tokens float64) error {
	if tokens > tb.MaxTokens {
		return fmt.Errorf("can't reserve more than %f tokens:%w", tb.MaxTokens, errOutOfBoundsRequest)
	}

	tb.refill()
	if tokens > tb.Tokens {
		return ErrTooManyRequests{tb.howMuchToWaitFor(tokens)}
	}

	tb.Tokens -= tokens

	return nil
}

func (tb *TokenBucket) SortValue() string {
	return fmt.Sprintf("%v", tb.ExpireAt())
}

func (tb *TokenBucket) ExpireAt() time.Time {
	durationSeconds := tb.howMuchToWaitFor(tb.MaxTokens)
	return tb.now().Add(durationSeconds)
}
