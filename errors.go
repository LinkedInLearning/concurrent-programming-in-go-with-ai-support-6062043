package ratelimiter

import "errors"

var (
	ErrRateLimiterStopped = errors.New("rate limiter has been stopped")
)
