package client_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/concurrent-programming-in-go/ratelimiter"
)

func ExampleTokenBucket_Allow() {
	tb := ratelimiter.NewTokenBucket(3, 100*time.Millisecond)
	defer tb.Stop()

	for i := 0; i < 5; i++ {
		if tb.Allow() {
			fmt.Printf("Request %d: Allowed\n", i+1)
		} else {
			fmt.Printf("Request %d: Rate limited\n", i+1)
		}
	}
}

func ExampleTokenBucket_Wait() {
	tb := ratelimiter.NewTokenBucket(2, 50*time.Millisecond)
	defer tb.Stop()

	ctx := context.Background()

	for i := 0; i < 3; i++ {
		start := time.Now()
		err := tb.Wait(ctx)
		if err != nil {
			log.Printf("Error waiting for token: %v", err)
			continue
		}
		duration := time.Since(start)
		fmt.Printf("Request %d processed after waiting %v\n", i+1, duration.Round(time.Millisecond))
	}
}

func ExampleTokenBucket_withHTTPServer() {
	tb := ratelimiter.NewTokenBucket(10, 100*time.Millisecond)
	defer tb.Stop()

	handleRequest := func(userID string) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		if err := tb.Wait(ctx); err != nil {
			fmt.Printf("Rate limit exceeded for user %s: %v\n", userID, err)
			return
		}

		fmt.Printf("Processing request for user %s\n", userID)
	}

	handleRequest("user1")
	handleRequest("user2")
	handleRequest("user3")
}
