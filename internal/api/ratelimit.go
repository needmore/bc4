package api

import (
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// LowRemainingThreshold is the threshold below which we proactively slow down requests
const LowRemainingThreshold = 5

// RateLimitInfo contains rate limit information parsed from API response headers
type RateLimitInfo struct {
	Limit      int       // Total requests allowed per window
	Remaining  int       // Requests remaining in current window
	Reset      time.Time // When the window resets
	RetryAfter int       // Seconds to wait (from 429 responses)
}

// ParseRateLimitHeaders extracts rate limit information from HTTP response headers.
// It handles the following headers:
//   - Retry-After: seconds to wait (on 429 responses)
//   - X-RateLimit-Limit: total requests allowed per window
//   - X-RateLimit-Remaining: requests remaining in current window
//   - X-RateLimit-Reset: Unix timestamp when the window resets
//
// Returns nil if no rate limit headers are present.
func ParseRateLimitHeaders(headers http.Header) *RateLimitInfo {
	info := &RateLimitInfo{}
	hasData := false

	// Parse Retry-After header (usually on 429 responses)
	if retryAfter := headers.Get("Retry-After"); retryAfter != "" {
		if seconds, err := strconv.Atoi(retryAfter); err == nil {
			info.RetryAfter = seconds
			hasData = true
		}
	}

	// Parse X-RateLimit-Limit
	if limit := headers.Get("X-RateLimit-Limit"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil {
			info.Limit = val
			hasData = true
		}
	}

	// Parse X-RateLimit-Remaining
	if remaining := headers.Get("X-RateLimit-Remaining"); remaining != "" {
		if val, err := strconv.Atoi(remaining); err == nil {
			info.Remaining = val
			hasData = true
		}
	}

	// Parse X-RateLimit-Reset (Unix timestamp)
	if reset := headers.Get("X-RateLimit-Reset"); reset != "" {
		if timestamp, err := strconv.ParseInt(reset, 10, 64); err == nil {
			info.Reset = time.Unix(timestamp, 0)
			hasData = true
		}
	}

	if !hasData {
		return nil
	}

	return info
}

// RateLimiter implements a token bucket algorithm for rate limiting
// Basecamp allows 50 requests per 10 seconds
type RateLimiter struct {
	mu         sync.Mutex
	tokens     int
	maxTokens  int
	refillRate time.Duration
	lastRefill time.Time
	debug      bool // Enable debug logging
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

// Wait blocks until a token is available
func (rl *RateLimiter) Wait() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Refill tokens based on time elapsed
	rl.refill()

	// If we have tokens, consume one and return
	if rl.tokens > 0 {
		rl.tokens--
		return
	}

	// Calculate wait time until next token
	timeSinceLastRefill := time.Since(rl.lastRefill)
	waitTime := rl.refillRate - timeSinceLastRefill

	if waitTime > 0 {
		rl.mu.Unlock()
		time.Sleep(waitTime)
		rl.mu.Lock()
		rl.refill()
		rl.tokens--
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

// SetDebug enables or disables debug logging
func (rl *RateLimiter) SetDebug(enabled bool) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.debug = enabled
}

// UpdateFromHeaders updates the rate limiter state based on API response headers.
// This allows dynamic adjustment based on actual server-reported limits.
func (rl *RateLimiter) UpdateFromHeaders(info *RateLimitInfo) {
	if info == nil {
		return
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Update tokens based on remaining count from headers
	if info.Remaining > 0 {
		// Use the actual remaining count from the server
		rl.tokens = info.Remaining

		if rl.debug && info.Remaining <= LowRemainingThreshold {
			log.Printf("[ratelimit] Low remaining requests: %d/%d", info.Remaining, info.Limit)
		}
	}

	// Update max tokens if limit is provided and different
	if info.Limit > 0 && info.Limit != rl.maxTokens {
		rl.maxTokens = info.Limit
		// Recalculate refill rate based on 10 second window (Basecamp default)
		rl.refillRate = (10 * time.Second) / time.Duration(info.Limit)

		if rl.debug {
			log.Printf("[ratelimit] Updated limit to %d requests per 10s", info.Limit)
		}
	}

	// If reset time is provided and in the future, adjust lastRefill
	if !info.Reset.IsZero() && info.Reset.After(time.Now()) {
		// The reset time tells us when tokens will be fully replenished
		rl.lastRefill = time.Now()
	}

	// If RetryAfter is set (from a 429), we should have 0 tokens
	if info.RetryAfter > 0 {
		rl.tokens = 0
		if rl.debug {
			log.Printf("[ratelimit] Rate limited, retry after %d seconds", info.RetryAfter)
		}
	}
}

// GetProactiveDelay returns a delay to apply when remaining requests are low.
// This helps avoid hitting 429 errors by slowing down proactively.
func (rl *RateLimiter) GetProactiveDelay(remaining int) time.Duration {
	if remaining > LowRemainingThreshold {
		return 0
	}

	// When remaining is low, add increasing delays
	// At 5 remaining: 200ms delay
	// At 1 remaining: 1s delay
	delayMs := (LowRemainingThreshold - remaining + 1) * 200
	return time.Duration(delayMs) * time.Millisecond
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
