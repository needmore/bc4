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
	Limit        int       // Total requests allowed per window
	Remaining    int       // Requests remaining in current window
	HasRemaining bool      // Whether Remaining was explicitly set (to distinguish 0 from unset)
	Reset        time.Time // When the window resets
	RetryAfter   int       // Seconds to wait (from 429 responses)
}

// ParseRateLimitHeaders extracts rate limit information from HTTP response headers.
// It handles the following headers:
//   - Retry-After: seconds to wait (on 429 responses)
//   - X-RateLimit-Limit: total requests allowed per window
//   - X-RateLimit-Remaining: requests remaining in current window
//   - X-RateLimit-Reset: Unix timestamp when the window resets
//
// Returns nil if no rate limit headers are present or all values are invalid.
// Negative values are rejected as invalid.
func ParseRateLimitHeaders(headers http.Header) *RateLimitInfo {
	info := &RateLimitInfo{}
	hasData := false

	// Parse Retry-After header (usually on 429 responses)
	if retryAfter := headers.Get("Retry-After"); retryAfter != "" {
		if seconds, err := strconv.Atoi(retryAfter); err == nil && seconds >= 0 {
			info.RetryAfter = seconds
			hasData = true
		} else if err != nil {
			log.Printf("[ratelimit] Warning: could not parse Retry-After header %q: %v", retryAfter, err)
		} else {
			log.Printf("[ratelimit] Warning: negative Retry-After value %d, ignoring", seconds)
		}
	}

	// Parse X-RateLimit-Limit
	if limit := headers.Get("X-RateLimit-Limit"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil && val >= 0 {
			info.Limit = val
			hasData = true
		} else if err != nil {
			log.Printf("[ratelimit] Warning: could not parse X-RateLimit-Limit header %q: %v", limit, err)
		} else {
			log.Printf("[ratelimit] Warning: negative X-RateLimit-Limit value %d, ignoring", val)
		}
	}

	// Parse X-RateLimit-Remaining
	if remaining := headers.Get("X-RateLimit-Remaining"); remaining != "" {
		if val, err := strconv.Atoi(remaining); err == nil && val >= 0 {
			info.Remaining = val
			info.HasRemaining = true
			hasData = true
		} else if err != nil {
			log.Printf("[ratelimit] Warning: could not parse X-RateLimit-Remaining header %q: %v", remaining, err)
		} else {
			log.Printf("[ratelimit] Warning: negative X-RateLimit-Remaining value %d, ignoring", val)
		}
	}

	// Parse X-RateLimit-Reset (Unix timestamp)
	if reset := headers.Get("X-RateLimit-Reset"); reset != "" {
		if timestamp, err := strconv.ParseInt(reset, 10, 64); err == nil && timestamp >= 0 {
			info.Reset = time.Unix(timestamp, 0)
			hasData = true
		} else if err != nil {
			log.Printf("[ratelimit] Warning: could not parse X-RateLimit-Reset header %q: %v", reset, err)
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

	// Update max tokens if limit is provided and different
	if info.Limit > 0 && info.Limit != rl.maxTokens {
		rl.maxTokens = info.Limit
		// Recalculate refill rate based on 10 second window (Basecamp default)
		rl.refillRate = (10 * time.Second) / time.Duration(info.Limit)

		if rl.debug {
			log.Printf("[ratelimit] Updated limit to %d requests per 10s", info.Limit)
		}
	}

	// If RetryAfter is set (from a 429), set tokens to 0 and adjust timing
	// This takes precedence over Remaining since it's the server's explicit instruction
	if info.RetryAfter > 0 {
		rl.tokens = 0
		// Use the RetryAfter value to inform when we can resume
		// Set lastRefill to now minus the full refill duration plus RetryAfter
		// This ensures Wait() will block for approximately RetryAfter seconds
		rl.lastRefill = time.Now().Add(-time.Duration(rl.maxTokens) * rl.refillRate).Add(time.Duration(info.RetryAfter) * time.Second)
		if rl.debug {
			log.Printf("[ratelimit] Rate limited, retry after %d seconds", info.RetryAfter)
		}
		return // Don't process Remaining when we have RetryAfter
	}

	// Update tokens based on remaining count from headers
	// Use HasRemaining to distinguish between "0 remaining" and "not provided"
	if info.HasRemaining {
		rl.tokens = info.Remaining

		if rl.debug {
			if info.Remaining == 0 {
				log.Printf("[ratelimit] No requests remaining, waiting for refill")
			} else if info.Remaining <= LowRemainingThreshold {
				log.Printf("[ratelimit] Low remaining requests: %d/%d", info.Remaining, info.Limit)
			}
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
	// At 0 remaining: 1.2s delay
	if remaining < 0 {
		remaining = 0
	}
	delayMs := (LowRemainingThreshold - remaining + 1) * 200
	return time.Duration(delayMs) * time.Millisecond
}

// Debug returns whether debug logging is enabled (thread-safe)
func (rl *RateLimiter) Debug() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.debug
}

// Tokens returns the current token count (thread-safe, for testing)
func (rl *RateLimiter) Tokens() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.tokens
}
