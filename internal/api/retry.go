package api

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/needmore/bc4/internal/errors"
)

// RetryConfig defines the configuration for retry behavior
type RetryConfig struct {
	MaxRetries           int           // Default: 3
	InitialBackoff       time.Duration // Default: 1s
	MaxBackoff           time.Duration // Default: 60s
	Multiplier           float64       // Default: 2.0
	RetryableStatusCodes []int         // Default: [429, 500, 502, 503, 504]
}

// DefaultRetryConfig returns the default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:           3,
		InitialBackoff:       1 * time.Second,
		MaxBackoff:           60 * time.Second,
		Multiplier:           2.0,
		RetryableStatusCodes: []int{429, 500, 502, 503, 504},
	}
}

// RetryableTransport wraps an http.RoundTripper to add retry logic
type RetryableTransport struct {
	Base   http.RoundTripper
	Config RetryConfig
}

// NewRetryableTransport creates a new retryable transport
func NewRetryableTransport(base http.RoundTripper, config RetryConfig) *RetryableTransport {
	if base == nil {
		base = http.DefaultTransport
	}
	return &RetryableTransport{
		Base:   base,
		Config: config,
	}
}

// RoundTrip implements http.RoundTripper with retry logic
func (rt *RetryableTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Buffer the request body if present so we can replay it on retries
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
		req.Body.Close()
	}

	var lastErr error
	var resp *http.Response

	for attempt := 0; attempt <= rt.Config.MaxRetries; attempt++ {
		// Reset body for retry attempts
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		// Make the request
		resp, lastErr = rt.Base.RoundTrip(req)

		// If we got an error (network error), retry if we haven't exhausted attempts
		if lastErr != nil {
			if attempt == rt.Config.MaxRetries {
				break
			}
			// Calculate backoff and retry
			backoff := rt.calculateBackoff(attempt, nil)
			time.Sleep(backoff)
			continue
		}

		// If no error and response is not retryable, return immediately
		if !rt.isRetryableStatus(resp.StatusCode) {
			return resp, nil
		}

		// Close the response body before retrying
		if resp.Body != nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}

		// If this was the last attempt, break and return the response
		if attempt == rt.Config.MaxRetries {
			break
		}

		// Calculate backoff duration
		backoff := rt.calculateBackoff(attempt, resp)

		// Wait before retrying
		time.Sleep(backoff)
	}

	// All retries exhausted
	if lastErr != nil {
		return nil, errors.NewRetryExhaustedError(rt.Config.MaxRetries+1, lastErr)
	}

	// Return the last response even though it's an error status
	// The caller will handle the HTTP error status
	return resp, nil
}

// isRetryableStatus checks if a status code should trigger a retry
func (rt *RetryableTransport) isRetryableStatus(statusCode int) bool {
	for _, code := range rt.Config.RetryableStatusCodes {
		if code == statusCode {
			return true
		}
	}
	return false
}

// calculateBackoff calculates the backoff duration for a retry attempt
func (rt *RetryableTransport) calculateBackoff(attempt int, resp *http.Response) time.Duration {
	// Check for Retry-After header
	if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
		if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
			// Try to parse as seconds
			if seconds, err := strconv.Atoi(retryAfter); err == nil {
				duration := time.Duration(seconds) * time.Second
				// Cap at max backoff
				if duration > rt.Config.MaxBackoff {
					return rt.Config.MaxBackoff
				}
				return duration
			}
			// Could also parse HTTP-date format here if needed
		}
	}

	// Use exponential backoff
	backoff := float64(rt.Config.InitialBackoff) * math.Pow(rt.Config.Multiplier, float64(attempt))
	duration := time.Duration(backoff)

	// Cap at max backoff
	if duration > rt.Config.MaxBackoff {
		return rt.Config.MaxBackoff
	}

	return duration
}
