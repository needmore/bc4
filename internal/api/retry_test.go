package api

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/needmore/bc4/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTransport is a mock http.RoundTripper for testing
type mockTransport struct {
	responses []*http.Response
	errors    []error
	callCount int
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.callCount >= len(m.responses) {
		m.callCount++
		if len(m.errors) > 0 {
			return nil, m.errors[len(m.errors)-1]
		}
		return m.responses[len(m.responses)-1], nil
	}

	idx := m.callCount
	m.callCount++

	if len(m.errors) > idx && m.errors[idx] != nil {
		return nil, m.errors[idx]
	}

	return m.responses[idx], nil
}

func newMockResponse(statusCode int, body string, headers map[string]string) *http.Response {
	resp := &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Header:     make(http.Header),
	}
	for k, v := range headers {
		resp.Header.Set(k, v)
	}
	return resp
}

func TestRetryableTransport_SuccessFirstAttempt(t *testing.T) {
	mock := &mockTransport{
		responses: []*http.Response{
			newMockResponse(200, "success", nil),
		},
	}

	config := DefaultRetryConfig()
	rt := NewRetryableTransport(mock, config)

	req, err := http.NewRequest("GET", "http://example.com", nil)
	require.NoError(t, err)

	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, 1, mock.callCount)
}

func TestRetryableTransport_Retry429(t *testing.T) {
	mock := &mockTransport{
		responses: []*http.Response{
			newMockResponse(429, "rate limited", nil),
			newMockResponse(429, "rate limited", nil),
			newMockResponse(200, "success", nil),
		},
	}

	config := DefaultRetryConfig()
	config.InitialBackoff = 10 * time.Millisecond // Speed up test
	rt := NewRetryableTransport(mock, config)

	req, err := http.NewRequest("GET", "http://example.com", nil)
	require.NoError(t, err)

	start := time.Now()
	resp, err := rt.RoundTrip(req)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, 3, mock.callCount)

	// Should have some backoff delay (at least 10ms + 20ms)
	assert.Greater(t, elapsed, 30*time.Millisecond)
}

func TestRetryableTransport_Retry5xxErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"500 Internal Server Error", 500},
		{"502 Bad Gateway", 502},
		{"503 Service Unavailable", 503},
		{"504 Gateway Timeout", 504},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockTransport{
				responses: []*http.Response{
					newMockResponse(tt.statusCode, "error", nil),
					newMockResponse(tt.statusCode, "error", nil),
					newMockResponse(200, "success", nil),
				},
			}

			config := DefaultRetryConfig()
			config.InitialBackoff = 10 * time.Millisecond
			rt := NewRetryableTransport(mock, config)

			req, err := http.NewRequest("GET", "http://example.com", nil)
			require.NoError(t, err)

			resp, err := rt.RoundTrip(req)
			require.NoError(t, err)
			assert.Equal(t, 200, resp.StatusCode)
			assert.Equal(t, 3, mock.callCount)
		})
	}
}

func TestRetryableTransport_RetryAfterHeader(t *testing.T) {
	mock := &mockTransport{
		responses: []*http.Response{
			newMockResponse(429, "rate limited", map[string]string{"Retry-After": "1"}),
			newMockResponse(200, "success", nil),
		},
	}

	config := DefaultRetryConfig()
	rt := NewRetryableTransport(mock, config)

	req, err := http.NewRequest("GET", "http://example.com", nil)
	require.NoError(t, err)

	start := time.Now()
	resp, err := rt.RoundTrip(req)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, 2, mock.callCount)

	// Should wait at least 1 second based on Retry-After header
	assert.Greater(t, elapsed, 1*time.Second)
}

func TestRetryableTransport_MaxRetriesExhausted(t *testing.T) {
	mock := &mockTransport{
		responses: []*http.Response{
			newMockResponse(500, "error", nil),
			newMockResponse(500, "error", nil),
			newMockResponse(500, "error", nil),
			newMockResponse(500, "error", nil), // All attempts fail
		},
	}

	config := DefaultRetryConfig()
	config.MaxRetries = 3
	config.InitialBackoff = 10 * time.Millisecond
	rt := NewRetryableTransport(mock, config)

	req, err := http.NewRequest("GET", "http://example.com", nil)
	require.NoError(t, err)

	resp, err := rt.RoundTrip(req)
	require.NoError(t, err) // Transport returns response, not error
	assert.Equal(t, 500, resp.StatusCode)
	assert.Equal(t, 4, mock.callCount) // Initial + 3 retries
}

func TestRetryableTransport_NonRetryableStatusCodes(t *testing.T) {
	nonRetryableCodes := []int{400, 401, 403, 404, 422}

	for _, code := range nonRetryableCodes {
		t.Run(strconv.Itoa(code), func(t *testing.T) {
			mock := &mockTransport{
				responses: []*http.Response{
					newMockResponse(code, "error", nil),
				},
			}

			config := DefaultRetryConfig()
			rt := NewRetryableTransport(mock, config)

			req, err := http.NewRequest("GET", "http://example.com", nil)
			require.NoError(t, err)

			resp, err := rt.RoundTrip(req)
			require.NoError(t, err)
			assert.Equal(t, code, resp.StatusCode)
			assert.Equal(t, 1, mock.callCount) // No retry
		})
	}
}

func TestRetryableTransport_ExponentialBackoff(t *testing.T) {
	mock := &mockTransport{
		responses: []*http.Response{
			newMockResponse(500, "error", nil),
			newMockResponse(500, "error", nil),
			newMockResponse(500, "error", nil),
			newMockResponse(200, "success", nil),
		},
	}

	config := DefaultRetryConfig()
	config.InitialBackoff = 100 * time.Millisecond
	config.Multiplier = 2.0
	rt := NewRetryableTransport(mock, config)

	req, err := http.NewRequest("GET", "http://example.com", nil)
	require.NoError(t, err)

	start := time.Now()
	resp, err := rt.RoundTrip(req)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, 4, mock.callCount)

	// Expected backoff: 100ms + 200ms + 400ms = 700ms
	// Allow some tolerance for test execution
	assert.Greater(t, elapsed, 600*time.Millisecond)
	assert.Less(t, elapsed, 1*time.Second)
}

func TestRetryableTransport_MaxBackoffCap(t *testing.T) {
	mock := &mockTransport{
		responses: []*http.Response{
			newMockResponse(429, "rate limited", map[string]string{"Retry-After": "120"}),
			newMockResponse(200, "success", nil),
		},
	}

	config := DefaultRetryConfig()
	config.MaxBackoff = 2 * time.Second
	rt := NewRetryableTransport(mock, config)

	req, err := http.NewRequest("GET", "http://example.com", nil)
	require.NoError(t, err)

	start := time.Now()
	resp, err := rt.RoundTrip(req)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Should be capped at MaxBackoff (2s) instead of 120s
	assert.Greater(t, elapsed, 2*time.Second)
	assert.Less(t, elapsed, 3*time.Second)
}

func TestRetryableTransport_RequestBodyBuffering(t *testing.T) {
	mock := &mockTransport{
		responses: []*http.Response{
			newMockResponse(500, "error", nil),
			newMockResponse(200, "success", nil),
		},
	}

	config := DefaultRetryConfig()
	config.InitialBackoff = 10 * time.Millisecond
	rt := NewRetryableTransport(mock, config)

	body := "test request body"
	req, err := http.NewRequest("POST", "http://example.com", bytes.NewBufferString(body))
	require.NoError(t, err)

	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, 2, mock.callCount)
}

func TestRetryableTransport_NetworkError(t *testing.T) {
	mock := &mockTransport{
		responses: []*http.Response{
			nil,
			nil,
			nil,
			nil, // All network errors
		},
		errors: []error{
			assert.AnError,
			assert.AnError,
			assert.AnError,
			assert.AnError, // All attempts fail
		},
	}

	config := DefaultRetryConfig()
	config.InitialBackoff = 10 * time.Millisecond
	rt := NewRetryableTransport(mock, config)

	req, err := http.NewRequest("GET", "http://example.com", nil)
	require.NoError(t, err)

	_, err = rt.RoundTrip(req)
	require.Error(t, err)

	// Should be a RetryExhaustedError
	var retryErr *errors.RetryExhaustedError
	assert.True(t, errors.IsRetryExhaustedError(err))
	require.ErrorAs(t, err, &retryErr)
	assert.Equal(t, 4, retryErr.Attempts) // Initial + 3 retries
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 1*time.Second, config.InitialBackoff)
	assert.Equal(t, 60*time.Second, config.MaxBackoff)
	assert.Equal(t, 2.0, config.Multiplier)
	assert.ElementsMatch(t, []int{429, 500, 502, 503, 504}, config.RetryableStatusCodes)
}

func TestRetryableTransport_CalculateBackoff(t *testing.T) {
	config := DefaultRetryConfig()
	config.InitialBackoff = 1 * time.Second
	config.Multiplier = 2.0
	config.MaxBackoff = 60 * time.Second

	rt := NewRetryableTransport(nil, config)

	tests := []struct {
		name     string
		attempt  int
		resp     *http.Response
		expected time.Duration
	}{
		{
			name:     "first attempt",
			attempt:  0,
			resp:     nil,
			expected: 1 * time.Second,
		},
		{
			name:     "second attempt",
			attempt:  1,
			resp:     nil,
			expected: 2 * time.Second,
		},
		{
			name:     "third attempt",
			attempt:  2,
			resp:     nil,
			expected: 4 * time.Second,
		},
		{
			name:     "with Retry-After header",
			attempt:  0,
			resp:     newMockResponse(429, "", map[string]string{"Retry-After": "5"}),
			expected: 5 * time.Second,
		},
		{
			name:     "Retry-After exceeds max",
			attempt:  0,
			resp:     newMockResponse(429, "", map[string]string{"Retry-After": "120"}),
			expected: 60 * time.Second, // Capped at MaxBackoff
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backoff := rt.calculateBackoff(tt.attempt, tt.resp)
			assert.Equal(t, tt.expected, backoff)
		})
	}
}

func TestNewClientWithRetryConfig(t *testing.T) {
	config := RetryConfig{
		MaxRetries:           5,
		InitialBackoff:       2 * time.Second,
		MaxBackoff:           120 * time.Second,
		Multiplier:           3.0,
		RetryableStatusCodes: []int{429, 500},
	}

	client := NewClientWithRetryConfig("account123", "token456", config)

	assert.NotNil(t, client)
	assert.Equal(t, "account123", client.accountID)
	assert.Equal(t, "token456", client.accessToken)
	assert.NotNil(t, client.httpClient)
	assert.NotNil(t, client.httpClient.Transport)

	// Verify transport is RetryableTransport
	rt, ok := client.httpClient.Transport.(*RetryableTransport)
	require.True(t, ok, "Transport should be RetryableTransport")
	assert.Equal(t, config.MaxRetries, rt.Config.MaxRetries)
	assert.Equal(t, config.InitialBackoff, rt.Config.InitialBackoff)
}

func TestNewClient_UsesDefaultRetryConfig(t *testing.T) {
	client := NewClient("account123", "token456")

	assert.NotNil(t, client)
	assert.NotNil(t, client.httpClient)
	assert.NotNil(t, client.httpClient.Transport)

	rt, ok := client.httpClient.Transport.(*RetryableTransport)
	require.True(t, ok, "Transport should be RetryableTransport")

	// Verify default config
	assert.Equal(t, 3, rt.Config.MaxRetries)
	assert.Equal(t, 1*time.Second, rt.Config.InitialBackoff)
	assert.Equal(t, 60*time.Second, rt.Config.MaxBackoff)
}
