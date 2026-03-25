package openrouter

import (
	"sync"
	"time"
)

type RateLimiter struct {
	maxRequests int
	window      time.Duration
	requests    []time.Time
	mutex       sync.Mutex
}

func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		maxRequests: maxRequests,
		window:      window,
		requests:    make([]time.Time, 0),
	}
}

func (rl *RateLimiter) Allow() bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	validRequests := make([]time.Time, 0)
	for _, reqTime := range rl.requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	rl.requests = validRequests

	if len(rl.requests) >= rl.maxRequests {
		return false
	}

	rl.requests = append(rl.requests, now)
	return true
}

func (rl *RateLimiter) GetCurrentUsage() int {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	count := 0
	for _, reqTime := range rl.requests {
		if reqTime.After(cutoff) {
			count++
		}
	}

	return count
}

func (rl *RateLimiter) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"max_requests":    rl.maxRequests,
		"window_seconds":  rl.window.Seconds(),
		"current_usage":   rl.GetCurrentUsage(),
		"remaining":       rl.maxRequests - rl.GetCurrentUsage(),
	}
}

func (rl *RateLimiter) Reset() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	rl.requests = make([]time.Time, 0)
}
