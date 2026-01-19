package api

import (
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRateLimitHeaders_RetryAfter(t *testing.T) {
	headers := http.Header{}
	headers.Set("Retry-After", "30")

	info := ParseRateLimitHeaders(headers)
	require.NotNil(t, info)
	assert.Equal(t, 30, info.RetryAfter)
}

func TestParseRateLimitHeaders_AllHeaders(t *testing.T) {
	headers := http.Header{}
	headers.Set("X-RateLimit-Limit", "50")
	headers.Set("X-RateLimit-Remaining", "45")
	headers.Set("X-RateLimit-Reset", "1609459200")

	info := ParseRateLimitHeaders(headers)
	require.NotNil(t, info)
	assert.Equal(t, 50, info.Limit)
	assert.Equal(t, 45, info.Remaining)
	assert.True(t, info.HasRemaining)
	assert.Equal(t, time.Unix(1609459200, 0), info.Reset)
}

func TestParseRateLimitHeaders_NoHeaders(t *testing.T) {
	headers := http.Header{}

	info := ParseRateLimitHeaders(headers)
	assert.Nil(t, info)
}

func TestParseRateLimitHeaders_InvalidValues(t *testing.T) {
	headers := http.Header{}
	headers.Set("X-RateLimit-Limit", "invalid")
	headers.Set("X-RateLimit-Remaining", "not-a-number")
	headers.Set("Retry-After", "abc")

	info := ParseRateLimitHeaders(headers)
	assert.Nil(t, info) // No valid data parsed
}

func TestParseRateLimitHeaders_NegativeValues(t *testing.T) {
	headers := http.Header{}
	headers.Set("X-RateLimit-Remaining", "-5")
	headers.Set("X-RateLimit-Limit", "-1")
	headers.Set("Retry-After", "-10")

	info := ParseRateLimitHeaders(headers)
	assert.Nil(t, info) // Negative values are rejected
}

func TestParseRateLimitHeaders_PartialHeaders(t *testing.T) {
	headers := http.Header{}
	headers.Set("X-RateLimit-Remaining", "10")

	info := ParseRateLimitHeaders(headers)
	require.NotNil(t, info)
	assert.Equal(t, 0, info.Limit)      // Not provided
	assert.Equal(t, 10, info.Remaining) // Parsed
	assert.True(t, info.HasRemaining)   // Was explicitly set
	assert.Equal(t, 0, info.RetryAfter) // Not provided
	assert.True(t, info.Reset.IsZero()) // Not provided
}

func TestParseRateLimitHeaders_RemainingZero(t *testing.T) {
	headers := http.Header{}
	headers.Set("X-RateLimit-Remaining", "0")
	headers.Set("X-RateLimit-Limit", "50")

	info := ParseRateLimitHeaders(headers)
	require.NotNil(t, info)
	assert.Equal(t, 0, info.Remaining)
	assert.True(t, info.HasRemaining, "HasRemaining should be true when Remaining is explicitly 0")
	assert.Equal(t, 50, info.Limit)
}

func TestRateLimiter_UpdateFromHeaders_Remaining(t *testing.T) {
	rl := NewRateLimiter(50, 10*time.Second)

	info := &RateLimitInfo{
		Remaining:    25,
		HasRemaining: true,
	}

	rl.UpdateFromHeaders(info)

	// Verify tokens were updated using thread-safe getter
	assert.Equal(t, 25, rl.Tokens())
}

func TestRateLimiter_UpdateFromHeaders_RemainingZero(t *testing.T) {
	rl := NewRateLimiter(50, 10*time.Second)

	info := &RateLimitInfo{
		Remaining:    0,
		HasRemaining: true,
		Limit:        50,
	}

	rl.UpdateFromHeaders(info)

	// Verify tokens were set to 0
	assert.Equal(t, 0, rl.Tokens(), "tokens should be set to 0 when Remaining is 0")
}

func TestRateLimiter_UpdateFromHeaders_Limit(t *testing.T) {
	rl := NewRateLimiter(50, 10*time.Second)

	info := &RateLimitInfo{
		Limit:        100,
		Remaining:    80,
		HasRemaining: true,
	}

	rl.UpdateFromHeaders(info)

	// Verify maxTokens and tokens were updated
	rl.mu.Lock()
	assert.Equal(t, 100, rl.maxTokens)
	assert.Equal(t, 80, rl.tokens)
	// Verify refill rate was recalculated: 10s / 100 = 100ms
	assert.Equal(t, 100*time.Millisecond, rl.refillRate)
	rl.mu.Unlock()
}

func TestRateLimiter_UpdateFromHeaders_RetryAfter(t *testing.T) {
	rl := NewRateLimiter(50, 10*time.Second)

	info := &RateLimitInfo{
		RetryAfter: 30,
	}

	rl.UpdateFromHeaders(info)

	// Verify tokens were set to 0
	assert.Equal(t, 0, rl.Tokens())
}

func TestRateLimiter_UpdateFromHeaders_RetryAfterOverridesRemaining(t *testing.T) {
	rl := NewRateLimiter(50, 10*time.Second)

	info := &RateLimitInfo{
		Remaining:    25,
		HasRemaining: true,
		RetryAfter:   30,
	}

	rl.UpdateFromHeaders(info)

	// RetryAfter should take precedence - tokens should be 0
	assert.Equal(t, 0, rl.Tokens(), "RetryAfter should override Remaining")
}

func TestRateLimiter_UpdateFromHeaders_Nil(t *testing.T) {
	rl := NewRateLimiter(50, 10*time.Second)
	originalTokens := rl.Tokens()

	rl.UpdateFromHeaders(nil)

	// Verify nothing changed
	assert.Equal(t, originalTokens, rl.Tokens())
}

func TestRateLimiter_UpdateFromHeaders_NoHasRemaining(t *testing.T) {
	rl := NewRateLimiter(50, 10*time.Second)

	// Simulate parsing where Remaining header was not present
	info := &RateLimitInfo{
		Remaining:    0, // Zero value, but not explicitly set
		HasRemaining: false,
		Limit:        50,
	}

	rl.UpdateFromHeaders(info)

	// Tokens should NOT be updated to 0 when HasRemaining is false
	assert.Equal(t, 50, rl.Tokens(), "tokens should not change when HasRemaining is false")
}

func TestRateLimiter_GetProactiveDelay(t *testing.T) {
	rl := NewRateLimiter(50, 10*time.Second)

	tests := []struct {
		remaining   int
		expectDelay bool
		minDelay    time.Duration
	}{
		{remaining: 10, expectDelay: false, minDelay: 0},
		{remaining: 6, expectDelay: false, minDelay: 0},
		{remaining: 5, expectDelay: true, minDelay: 200 * time.Millisecond},
		{remaining: 3, expectDelay: true, minDelay: 600 * time.Millisecond},
		{remaining: 1, expectDelay: true, minDelay: 1000 * time.Millisecond},
		{remaining: 0, expectDelay: true, minDelay: 1200 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(strconv.Itoa(tt.remaining), func(t *testing.T) {
			delay := rl.GetProactiveDelay(tt.remaining)
			if tt.expectDelay {
				assert.GreaterOrEqual(t, delay, tt.minDelay)
			} else {
				assert.Equal(t, time.Duration(0), delay)
			}
		})
	}
}

func TestRateLimiter_GetProactiveDelay_NegativeRemaining(t *testing.T) {
	rl := NewRateLimiter(50, 10*time.Second)

	// Negative values should be treated as 0
	delay := rl.GetProactiveDelay(-5)
	assert.Equal(t, 1200*time.Millisecond, delay, "negative remaining should be treated as 0")
}

func TestRateLimiter_ThreadSafety(t *testing.T) {
	rl := NewRateLimiter(50, 10*time.Second)

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent UpdateFromHeaders calls
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			info := &RateLimitInfo{
				Remaining:    i % 50,
				HasRemaining: true,
				Limit:        50,
			}
			rl.UpdateFromHeaders(info)
		}(i)
	}

	// Concurrent TryAcquire calls
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rl.TryAcquire()
		}()
	}

	// Concurrent Reset calls
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rl.Reset()
		}()
	}

	// Concurrent Debug/SetDebug calls
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rl.SetDebug(true)
			_ = rl.Debug()
			rl.SetDebug(false)
		}()
	}

	wg.Wait()
	// If we get here without deadlock or panic, the test passes
}

func TestRateLimiter_SetDebug(t *testing.T) {
	rl := NewRateLimiter(50, 10*time.Second)

	// Use thread-safe getter
	assert.False(t, rl.Debug())

	rl.SetDebug(true)
	assert.True(t, rl.Debug())

	rl.SetDebug(false)
	assert.False(t, rl.Debug())
}

func TestRateLimiter_Wait(t *testing.T) {
	// Create a rate limiter with only 2 tokens
	rl := NewRateLimiter(2, 100*time.Millisecond)

	// First two waits should be instant
	start := time.Now()
	rl.Wait()
	rl.Wait()
	elapsed := time.Since(start)
	assert.Less(t, elapsed, 50*time.Millisecond)

	// Third wait should block until refill
	start = time.Now()
	rl.Wait()
	elapsed = time.Since(start)
	assert.Greater(t, elapsed, 40*time.Millisecond) // Should wait for refill
}

func TestRateLimiter_TryAcquire(t *testing.T) {
	rl := NewRateLimiter(2, 1*time.Second)

	// First two should succeed
	assert.True(t, rl.TryAcquire())
	assert.True(t, rl.TryAcquire())

	// Third should fail (no tokens left)
	assert.False(t, rl.TryAcquire())
}

func TestGetRateLimiter_Singleton(t *testing.T) {
	rl1 := GetRateLimiter()
	rl2 := GetRateLimiter()

	assert.Same(t, rl1, rl2, "GetRateLimiter should return the same instance")
}

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(100, 20*time.Second)

	rl.mu.Lock()
	assert.Equal(t, 100, rl.maxTokens)
	assert.Equal(t, 100, rl.tokens)
	assert.Equal(t, 200*time.Millisecond, rl.refillRate) // 20s / 100 = 200ms per token
	rl.mu.Unlock()
}

func TestRateLimiter_Reset(t *testing.T) {
	rl := NewRateLimiter(50, 10*time.Second)

	// Consume some tokens
	rl.TryAcquire()
	rl.TryAcquire()
	rl.TryAcquire()

	assert.Equal(t, 47, rl.Tokens())

	// Reset
	rl.Reset()

	assert.Equal(t, 50, rl.Tokens())
}

func TestRateLimiter_Tokens(t *testing.T) {
	rl := NewRateLimiter(50, 10*time.Second)

	assert.Equal(t, 50, rl.Tokens())

	rl.TryAcquire()
	assert.Equal(t, 49, rl.Tokens())
}
