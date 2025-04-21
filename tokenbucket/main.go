package main

import (
	"context"
	"fmt"
	"time"
)

type RateLimiter struct {
	tokenBucket chan struct{}
	refillTicker *time.Ticker
	cancel context.CancelFunc
}

func NewRateLimiter(rateLimit int, refillDuration time.Duration) *RateLimiter {
	ctx, cancel := context.WithCancel(context.Background())
	rl := &RateLimiter{
		tokenBucket: make(chan struct{}, rateLimit),
		refillTicker: time.NewTicker(refillDuration),
		cancel: cancel,
	}
	for range rateLimit {
		rl.tokenBucket <- struct{}{}
	}
	go rl.refill(ctx)
	return rl
}

func (rl *RateLimiter) refill(ctx context.Context) {
	ticker := rl.refillTicker
	for {
		select {
		case <-ticker.C:
			select {
			case rl.tokenBucket <- struct{}{}:
			default:
			}
		case <-ctx.Done():
			return
		}
	}
}

func (rl *RateLimiter) IsAvailable() bool {
	select {
	case <-rl.tokenBucket:
		return true
	default:
		return false
	}
}

func (rl *RateLimiter) Stop() {
	rl.cancel()
	rl.refillTicker.Stop()
}

func main() {
	rl := NewRateLimiter(5, 500 * time.Millisecond)
	defer rl.Stop()
	
	numRequests := 10
	for i := range numRequests {
		if rl.IsAvailable() {
			fmt.Printf("Request %2d is sent successfully - %v\n", i + 1, time.Now().Format(time.RFC1123Z))
		} else {
			fmt.Printf("Request %2d is denied - %v\n", i + 1, time.Now().Format(time.RFC1123Z))
		}
		time.Sleep(200 * time.Millisecond)
	}

}
