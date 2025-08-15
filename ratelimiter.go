package ratelimiter

import (
	"context"
	"sync"
	"time"
)

const (
	DefaultBucketSize = 10
	DefaultRefillRate = time.Second
)

type TokenBucket struct {
	bucketSize int
	refillRate time.Duration
	tokens     chan struct{}
	ticker     *time.Ticker
	stopCh     chan struct{}
	mu         sync.RWMutex
	stopped    bool
}

func NewTokenBucket(bucketSize int, refillRate time.Duration) *TokenBucket {
	if bucketSize <= 0 {
		bucketSize = DefaultBucketSize
	}
	if refillRate <= 0 {
		refillRate = DefaultRefillRate
	}

	tb := &TokenBucket{
		bucketSize: bucketSize,
		refillRate: refillRate,
		tokens:     make(chan struct{}, bucketSize),
		ticker:     time.NewTicker(refillRate),
		stopCh:     make(chan struct{}),
	}

	for i := 0; i < bucketSize; i++ {
		tb.tokens <- struct{}{}
	}

	go tb.refillTokens()

	return tb
}

func (tb *TokenBucket) refillTokens() {
	for {
		select {
		case <-tb.ticker.C:
			select {
			case tb.tokens <- struct{}{}:
			default:
			}
		case <-tb.stopCh:
			return
		}
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.mu.RLock()
	if tb.stopped {
		tb.mu.RUnlock()
		return false
	}
	tb.mu.RUnlock()

	select {
	case <-tb.tokens:
		return true
	default:
		return false
	}
}

func (tb *TokenBucket) Wait(ctx context.Context) error {
	tb.mu.RLock()
	if tb.stopped {
		tb.mu.RUnlock()
		return ErrRateLimiterStopped
	}
	tb.mu.RUnlock()

	select {
	case <-tb.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (tb *TokenBucket) Stop() {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	if tb.stopped {
		return
	}

	tb.stopped = true
	tb.ticker.Stop()
	close(tb.stopCh)
}

func (tb *TokenBucket) AvailableTokens() int {
	return len(tb.tokens)
}

func (tb *TokenBucket) BucketSize() int {
	return tb.bucketSize
}

func (tb *TokenBucket) RefillRate() time.Duration {
	return tb.refillRate
}