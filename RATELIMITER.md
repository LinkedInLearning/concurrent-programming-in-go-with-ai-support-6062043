# Rate Limiter Package

A Go package implementing a token bucket rate limiter using channels and tickers for concurrent-safe rate limiting.

## Features

- **Token Bucket Algorithm**: Implements the classic token bucket rate limiting algorithm
- **Channel-based**: Uses Go channels for thread-safe token management
- **Ticker-based Refill**: Uses `time.Ticker` for consistent token refill intervals
- **Context Support**: Supports context cancellation for graceful timeouts
- **Concurrent Safe**: Safe for use across multiple goroutines
- **Configurable**: Customizable bucket size and refill rate

## Usage

### Basic Usage

```go
package main

import (
    "fmt"
    "time"
    
    "github.com/concurrent-programming-in-go/ratelimiter"
)

func main() {
    // Create a rate limiter with 10 tokens, refilling 1 token every 100ms
    tb := ratelimiter.NewTokenBucket(10, 100*time.Millisecond)
    defer tb.Stop()
    
    // Check if request is allowed (non-blocking)
    if tb.Allow() {
        fmt.Println("Request allowed")
    } else {
        fmt.Println("Request rate limited")
    }
}
```

### Waiting for Tokens

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/concurrent-programming-in-go/ratelimiter"
)

func main() {
    tb := ratelimiter.NewTokenBucket(5, 200*time.Millisecond)
    defer tb.Stop()
    
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()
    
    // Wait for a token (blocking with timeout)
    if err := tb.Wait(ctx); err != nil {
        fmt.Printf("Failed to get token: %v\n", err)
        return
    }
    
    fmt.Println("Token acquired, processing request")
}
```

## API Reference

### TokenBucket

#### Constructor

- `NewTokenBucket(bucketSize int, refillRate time.Duration) *TokenBucket`

#### Methods

- `Allow() bool` - Non-blocking token check
- `Wait(ctx context.Context) error` - Blocking wait with context
- `Stop()` - Cleanup resources
- `AvailableTokens() int` - Current token count
- `BucketSize() int` - Maximum bucket size
- `RefillRate() time.Duration` - Token refill interval

### Constants

- `DefaultBucketSize = 10`
- `DefaultRefillRate = time.Second`

### Errors

- `ErrRateLimiterStopped` - Returned when using a stopped rate limiter

## Testing

```bash
go test -v
go test -bench=.
```