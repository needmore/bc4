package errors

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		checkFn  func(error) bool
		expected bool
	}{
		{
			name:     "authentication error",
			err:      NewAuthenticationError(errors.New("token expired")),
			checkFn:  IsAuthenticationError,
			expected: true,
		},
		{
			name:     "not found error",
			err:      NewNotFoundError("project", "12345", nil),
			checkFn:  IsNotFoundError,
			expected: true,
		},
		{
			name:     "API error",
			err:      NewAPIError(http.StatusUnauthorized, "invalid token", nil),
			checkFn:  IsAPIError,
			expected: true,
		},
		{
			name:     "validation error",
			err:      NewValidationError("email", "invalid format", nil),
			checkFn:  IsValidationError,
			expected: true,
		},
		{
			name:     "network error",
			err:      NewNetworkError(errors.New("connection timeout")),
			checkFn:  IsNetworkError,
			expected: true,
		},
		{
			name:     "configuration error",
			err:      NewConfigurationError("missing OAuth credentials", nil),
			checkFn:  IsConfigurationError,
			expected: true,
		},
		{
			name:     "wrapped error",
			err:      fmt.Errorf("wrapped: %w", NewAuthenticationError(errors.New("inner"))),
			checkFn:  IsAuthenticationError,
			expected: true,
		},
		{
			name:     "non-matching error",
			err:      errors.New("generic error"),
			checkFn:  IsAuthenticationError,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.checkFn(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		contains []string
	}{
		{
			name: "authentication error",
			err:  NewAuthenticationError(errors.New("token expired")),
			contains: []string{
				"Authentication failed",
				"bc4 auth login",
			},
		},
		{
			name: "not found error with ID",
			err:  NewNotFoundError("Project", "12345", nil),
			contains: []string{
				"Project not found",
				"12345",
				"Check the ID for typos",
				"bc4 project list",
			},
		},
		{
			name: "API error 401",
			err:  NewAPIError(http.StatusUnauthorized, "invalid token", nil),
			contains: []string{
				"API request failed",
				"bc4 auth login",
			},
		},
		{
			name: "API error 403",
			err:  NewAPIError(http.StatusForbidden, "access denied", nil),
			contains: []string{
				"API request failed",
				"don't have permission",
			},
		},
		{
			name: "API error 429",
			err:  NewAPIError(http.StatusTooManyRequests, "rate limited", nil),
			contains: []string{
				"API request failed",
				"Rate limit exceeded",
			},
		},
		{
			name: "API error 500",
			err:  NewAPIError(http.StatusInternalServerError, "server error", nil),
			contains: []string{
				"API request failed",
				"Basecamp is experiencing issues",
				"status.basecamp.com",
			},
		},
		{
			name: "validation error",
			err:  NewValidationError("email", "must be a valid email address", nil),
			contains: []string{
				"Invalid input",
				"email",
				"Check your input",
			},
		},
		{
			name: "network error",
			err:  NewNetworkError(errors.New("connection refused")),
			contains: []string{
				"Network connection failed",
				"internet connection",
			},
		},
		{
			name: "configuration error",
			err:  NewConfigurationError("OAuth credentials missing", nil),
			contains: []string{
				"Configuration error",
				"OAuth credentials missing",
				"Run 'bc4' to start the setup wizard",
			},
		},
		{
			name: "OAuth not configured (string match)",
			err:  errors.New("OAuth credentials not configured"),
			contains: []string{
				"OAuth not configured",
				"Run 'bc4' to start the setup wizard",
			},
		},
		{
			name: "not authenticated (string match)",
			err:  errors.New("not authenticated. Run 'bc4' to set up"),
			contains: []string{
				"Not authenticated",
				"bc4 auth login",
			},
		},
		{
			name: "no account specified (string match)",
			err:  errors.New("no account specified and no default account set"),
			contains: []string{
				"No account selected",
				"bc4 account select",
				"--account",
			},
		},
		{
			name: "no project specified (string match)",
			err:  errors.New("no project specified and no default project set"),
			contains: []string{
				"No project selected",
				"bc4 project select",
				"--project",
			},
		},
		{
			name: "generic error",
			err:  errors.New("something went wrong"),
			contains: []string{
				"Error",
				"something went wrong",
			},
		},
		{
			name:     "nil error",
			err:      nil,
			contains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := FormatError(tt.err)

			if tt.err == nil {
				assert.Empty(t, output)
				return
			}

			for _, expected := range tt.contains {
				assert.Contains(t, output, expected, "Output should contain: %s", expected)
			}

			// Check that output is not empty for non-nil errors
			if tt.err != nil {
				assert.NotEmpty(t, output)
			}
		})
	}
}

func TestErrorUnwrap(t *testing.T) {
	innerErr := errors.New("inner error")

	tests := []struct {
		name string
		err  error
	}{
		{
			name: "authentication error unwrap",
			err:  NewAuthenticationError(innerErr),
		},
		{
			name: "not found error unwrap",
			err:  NewNotFoundError("resource", "id", innerErr),
		},
		{
			name: "API error unwrap",
			err:  NewAPIError(400, "bad request", innerErr),
		},
		{
			name: "validation error unwrap",
			err:  NewValidationError("field", "message", innerErr),
		},
		{
			name: "network error unwrap",
			err:  NewNetworkError(innerErr),
		},
		{
			name: "configuration error unwrap",
			err:  NewConfigurationError("message", innerErr),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unwrapped := errors.Unwrap(tt.err)
			assert.Equal(t, innerErr, unwrapped)
		})
	}
}

func TestNotFoundErrorString(t *testing.T) {
	tests := []struct {
		name     string
		resource string
		id       string
		expected string
	}{
		{
			name:     "with ID",
			resource: "project",
			id:       "12345",
			expected: "project with ID '12345' not found",
		},
		{
			name:     "without ID",
			resource: "projects",
			id:       "",
			expected: "projects not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewNotFoundError(tt.resource, tt.id, nil)
			assert.Equal(t, tt.expected, err.Error())
		})
	}
}

func TestAPIErrorString(t *testing.T) {
	err := NewAPIError(404, "Not Found", nil)
	expected := "API error (status 404): Not Found"
	assert.Equal(t, expected, err.Error())
}

func TestValidationErrorString(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		message  string
		expected string
	}{
		{
			name:     "with field",
			field:    "email",
			message:  "must be valid",
			expected: "invalid email: must be valid",
		},
		{
			name:     "without field",
			field:    "",
			message:  "input is required",
			expected: "input is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewValidationError(tt.field, tt.message, nil)
			assert.Equal(t, tt.expected, err.Error())
		})
	}
}

func TestFormatErrorStripsStyling(t *testing.T) {
	// Test that the output doesn't contain ANSI escape codes
	// when we check the content (lipgloss will handle actual rendering)
	err := NewAuthenticationError(errors.New("test"))
	output := FormatError(err)

	// Should not contain raw ANSI codes in our test
	assert.NotContains(t, output, "\x1b[")

	// But should contain the actual content
	assert.Contains(t, output, "Authentication failed")
}
