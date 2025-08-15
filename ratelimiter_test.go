package ratelimiter

import (
	"context"
	"testing"
	"time"
)

func TestNewTokenBucket(t *testing.T) {
	tb := NewTokenBucket(5, 100*time.Millisecond)
	defer tb.Stop()

	if tb.BucketSize() != 5 {
		t.Errorf("expected bucket size 5, got %d", tb.BucketSize())
	}

	if tb.RefillRate() != 100*time.Millisecond {
		t.Errorf("expected refill rate 100ms, got %v", tb.RefillRate())
	}

	if tb.AvailableTokens() != 5 {
		t.Errorf("expected 5 available tokens, got %d", tb.AvailableTokens())
	}
}

func TestNewTokenBucketDefaults(t *testing.T) {
	tb := NewTokenBucket(0, 0)
	defer tb.Stop()

	if tb.BucketSize() != DefaultBucketSize {
		t.Errorf("expected default bucket size %d, got %d", DefaultBucketSize, tb.BucketSize())
	}

	if tb.RefillRate() != DefaultRefillRate {
		t.Errorf("expected default refill rate %v, got %v", DefaultRefillRate, tb.RefillRate())
	}
}

func TestTokenBucketAllow(t *testing.T) {
	tb := NewTokenBucket(3, 100*time.Millisecond)
	defer tb.Stop()

	for i := 0; i < 3; i++ {
		if !tb.Allow() {
			t.Errorf("expected Allow() to return true for token %d", i+1)
		}
	}

	if tb.Allow() {
		t.Error("expected Allow() to return false when bucket is empty")
	}

	if tb.AvailableTokens() != 0 {
		t.Errorf("expected 0 available tokens, got %d", tb.AvailableTokens())
	}
}

func TestTokenBucketRefill(t *testing.T) {
	tb := NewTokenBucket(2, 50*time.Millisecond)
	defer tb.Stop()

	tb.Allow()
	tb.Allow()

	if tb.AvailableTokens() != 0 {
		t.Errorf("expected 0 available tokens, got %d", tb.AvailableTokens())
	}

	time.Sleep(60 * time.Millisecond)

	if tb.AvailableTokens() != 1 {
		t.Errorf("expected 1 available token after refill, got %d", tb.AvailableTokens())
	}

	if !tb.Allow() {
		t.Error("expected Allow() to return true after refill")
	}
}

func TestTokenBucketWait(t *testing.T) {
	tb := NewTokenBucket(1, 50*time.Millisecond)
	defer tb.Stop()

	tb.Allow()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := tb.Wait(ctx)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if duration < 40*time.Millisecond {
		t.Errorf("expected to wait at least 40ms, waited %v", duration)
	}
}

func TestTokenBucketWaitTimeout(t *testing.T) {
	tb := NewTokenBucket(1, 200*time.Millisecond)
	defer tb.Stop()

	tb.Allow()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := tb.Wait(ctx)

	if err != context.DeadlineExceeded {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
}

func TestTokenBucketStop(t *testing.T) {
	tb := NewTokenBucket(5, 100*time.Millisecond)

	if !tb.Allow() {
		t.Error("expected Allow() to return true before stop")
	}

	tb.Stop()

	if tb.Allow() {
		t.Error("expected Allow() to return false after stop")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := tb.Wait(ctx)
	if err != ErrRateLimiterStopped {
		t.Errorf("expected ErrRateLimiterStopped, got %v", err)
	}
}

func TestTokenBucketConcurrency(t *testing.T) {
	tb := NewTokenBucket(10, 10*time.Millisecond)
	defer tb.Stop()

	const numGoroutines = 20
	results := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()
			err := tb.Wait(ctx)
			results <- err == nil
		}()
	}

	successCount := 0
	for i := 0; i < numGoroutines; i++ {
		if <-results {
			successCount++
		}
	}

	if successCount < 10 {
		t.Errorf("expected at least 10 successful requests, got %d", successCount)
	}
}

func BenchmarkTokenBucketAllow(b *testing.B) {
	tb := NewTokenBucket(1000000, time.Nanosecond)
	defer tb.Stop()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tb.Allow()
		}
	})
}
