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

func TestParseRateLimitHeaders_PartialHeaders(t *testing.T) {
	headers := http.Header{}
	headers.Set("X-RateLimit-Remaining", "10")

	info := ParseRateLimitHeaders(headers)
	require.NotNil(t, info)
	assert.Equal(t, 0, info.Limit)      // Not provided
	assert.Equal(t, 10, info.Remaining) // Parsed
	assert.Equal(t, 0, info.RetryAfter) // Not provided
	assert.True(t, info.Reset.IsZero()) // Not provided
}

func TestRateLimiter_UpdateFromHeaders_Remaining(t *testing.T) {
	rl := NewRateLimiter(50, 10*time.Second)

	info := &RateLimitInfo{
		Remaining: 25,
	}

	rl.UpdateFromHeaders(info)

	// Verify tokens were updated
	rl.mu.Lock()
	assert.Equal(t, 25, rl.tokens)
	rl.mu.Unlock()
}

func TestRateLimiter_UpdateFromHeaders_Limit(t *testing.T) {
	rl := NewRateLimiter(50, 10*time.Second)

	info := &RateLimitInfo{
		Limit:     100,
		Remaining: 80,
	}

	rl.UpdateFromHeaders(info)

	// Verify maxTokens was updated
	rl.mu.Lock()
	assert.Equal(t, 100, rl.maxTokens)
	assert.Equal(t, 80, rl.tokens)
	rl.mu.Unlock()
}

func TestRateLimiter_UpdateFromHeaders_RetryAfter(t *testing.T) {
	rl := NewRateLimiter(50, 10*time.Second)

	info := &RateLimitInfo{
		RetryAfter: 30,
	}

	rl.UpdateFromHeaders(info)

	// Verify tokens were set to 0
	rl.mu.Lock()
	assert.Equal(t, 0, rl.tokens)
	rl.mu.Unlock()
}

func TestRateLimiter_UpdateFromHeaders_Nil(t *testing.T) {
	rl := NewRateLimiter(50, 10*time.Second)
	originalTokens := rl.tokens

	rl.UpdateFromHeaders(nil)

	// Verify nothing changed
	rl.mu.Lock()
	assert.Equal(t, originalTokens, rl.tokens)
	rl.mu.Unlock()
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
				Remaining: i % 50,
				Limit:     50,
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

	wg.Wait()
	// If we get here without deadlock or panic, the test passes
}

func TestRateLimiter_SetDebug(t *testing.T) {
	rl := NewRateLimiter(50, 10*time.Second)

	assert.False(t, rl.debug)

	rl.SetDebug(true)
	assert.True(t, rl.debug)

	rl.SetDebug(false)
	assert.False(t, rl.debug)
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

	assert.Equal(t, 100, rl.maxTokens)
	assert.Equal(t, 100, rl.tokens)
	assert.Equal(t, 200*time.Millisecond, rl.refillRate) // 20s / 100 = 200ms per token
}

func TestRateLimiter_Reset(t *testing.T) {
	rl := NewRateLimiter(50, 10*time.Second)

	// Consume some tokens
	rl.TryAcquire()
	rl.TryAcquire()
	rl.TryAcquire()

	rl.mu.Lock()
	assert.Equal(t, 47, rl.tokens)
	rl.mu.Unlock()

	// Reset
	rl.Reset()

	rl.mu.Lock()
	assert.Equal(t, 50, rl.tokens)
	rl.mu.Unlock()
}
