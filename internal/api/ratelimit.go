package api

import (
	"sync"
	"time"
)

// RateLimiter implements a token bucket algorithm for rate limiting
// Basecamp allows 50 requests per 10 seconds
type RateLimiter struct {
	mu         sync.Mutex
	tokens     int
	maxTokens  int
	refillRate time.Duration
	lastRefill time.Time
}

var (
	globalRateLimiter *RateLimiter
	once              sync.Once
)

// GetRateLimiter returns the global rate limiter instance
func GetRateLimiter() *RateLimiter {
	once.Do(func() {
		globalRateLimiter = NewRateLimiter(50, 10*time.Second)
	})
	return globalRateLimiter
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxTokens int, refillDuration time.Duration) *RateLimiter {
	return &RateLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillDuration / time.Duration(maxTokens),
		lastRefill: time.Now(),
	}
}

// Wait blocks until a token is available.
// Safe for concurrent use — re-checks token availability after waking.
func (rl *RateLimiter) Wait() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for {
		rl.refill()

		if rl.tokens > 0 {
			rl.tokens--
			return
		}

		// Calculate wait time until next token
		waitTime := rl.refillRate - time.Since(rl.lastRefill)
		if waitTime <= 0 {
			waitTime = rl.refillRate
		}

		// Release lock while sleeping, then re-acquire and re-check
		rl.mu.Unlock()
		time.Sleep(waitTime)
		rl.mu.Lock()
	}
}

// TryAcquire attempts to acquire a token without blocking
func (rl *RateLimiter) TryAcquire() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refill()

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

// refill adds tokens based on time elapsed (must be called with lock held)
func (rl *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)

	// Calculate how many tokens to add
	tokensToAdd := int(elapsed / rl.refillRate)

	if tokensToAdd > 0 {
		rl.tokens = min(rl.tokens+tokensToAdd, rl.maxTokens)
		rl.lastRefill = rl.lastRefill.Add(time.Duration(tokensToAdd) * rl.refillRate)
	}
}

// Reset resets the rate limiter to full capacity
func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.tokens = rl.maxTokens
	rl.lastRefill = time.Now()
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
