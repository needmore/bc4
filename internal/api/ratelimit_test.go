package api

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiter_Wait_BasicTokenConsumption(t *testing.T) {
	rl := NewRateLimiter(3, 3*time.Second) // 3 tokens, 1 per second refill

	// Should consume 3 tokens without blocking
	start := time.Now()
	rl.Wait()
	rl.Wait()
	rl.Wait()
	elapsed := time.Since(start)

	assert.Less(t, elapsed, 100*time.Millisecond, "consuming available tokens should not block")
}

func TestRateLimiter_Wait_BlocksWhenEmpty(t *testing.T) {
	rl := NewRateLimiter(1, 1*time.Second) // 1 token, refills every 1s

	// Consume the only token
	rl.Wait()

	// Next call should block until refill
	start := time.Now()
	rl.Wait()
	elapsed := time.Since(start)

	assert.GreaterOrEqual(t, elapsed, 500*time.Millisecond, "should block when no tokens available")
}

func TestRateLimiter_Wait_ConcurrentNeverNegative(t *testing.T) {
	// 5 tokens, fast refill. 20 goroutines competing.
	// After all complete, tokens should never have gone negative.
	rl := NewRateLimiter(5, 500*time.Millisecond)

	var wg sync.WaitGroup
	var acquired int64

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rl.Wait()
			atomic.AddInt64(&acquired, 1)
		}()
	}

	wg.Wait()

	assert.Equal(t, int64(20), acquired, "all goroutines should eventually acquire a token")

	// Verify tokens are non-negative
	rl.mu.Lock()
	assert.GreaterOrEqual(t, rl.tokens, 0, "tokens should never be negative")
	rl.mu.Unlock()
}

func TestRateLimiter_TryAcquire(t *testing.T) {
	rl := NewRateLimiter(2, 2*time.Second)

	assert.True(t, rl.TryAcquire(), "first acquire should succeed")
	assert.True(t, rl.TryAcquire(), "second acquire should succeed")
	assert.False(t, rl.TryAcquire(), "third acquire should fail (no tokens)")
}

func TestRateLimiter_Reset(t *testing.T) {
	rl := NewRateLimiter(3, 3*time.Second)

	// Drain all tokens
	rl.Wait()
	rl.Wait()
	rl.Wait()
	assert.False(t, rl.TryAcquire(), "should be empty after draining")

	// Reset should restore
	rl.Reset()
	assert.True(t, rl.TryAcquire(), "should have tokens after reset")
	assert.True(t, rl.TryAcquire(), "should have tokens after reset")
	assert.True(t, rl.TryAcquire(), "should have tokens after reset")
}

func TestRateLimiter_Refill(t *testing.T) {
	rl := NewRateLimiter(2, 200*time.Millisecond) // refill rate = 100ms per token

	// Drain all tokens
	rl.Wait()
	rl.Wait()
	assert.False(t, rl.TryAcquire(), "should be empty")

	// Wait for at least 1 token to refill
	time.Sleep(150 * time.Millisecond)
	assert.True(t, rl.TryAcquire(), "should have refilled at least 1 token")
}
