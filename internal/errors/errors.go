package errors

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Error types for categorizing errors
type (
	// AuthenticationError indicates an authentication failure
	AuthenticationError struct {
		Err error
	}

	// NotFoundError indicates a resource was not found
	NotFoundError struct {
		Resource string
		ID       string
		Err      error
	}

	// APIError represents an API failure
	APIError struct {
		StatusCode int
		Message    string
		Err        error
	}

	// ValidationError indicates invalid user input
	ValidationError struct {
		Field   string
		Message string
		Err     error
	}

	// NetworkError indicates a network connectivity issue
	NetworkError struct {
		Err error
	}

	// ConfigurationError indicates a configuration issue
	ConfigurationError struct {
		Message string
		Err     error
	}

	// RetryExhaustedError indicates all retry attempts failed
	RetryExhaustedError struct {
		Attempts int
		LastErr  error
	}
)

// Error implementations
func (e *AuthenticationError) Error() string {
	return fmt.Sprintf("authentication error: %v", e.Err)
}

func (e *AuthenticationError) Unwrap() error {
	return e.Err
}

func (e *NotFoundError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("%s with ID '%s' not found", e.Resource, e.ID)
	}
	return fmt.Sprintf("%s not found", e.Resource)
}

func (e *NotFoundError) Unwrap() error {
	return e.Err
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
}

func (e *APIError) Unwrap() error {
	return e.Err
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("invalid %s: %s", e.Field, e.Message)
	}
	return e.Message
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error: %v", e.Err)
}

func (e *NetworkError) Unwrap() error {
	return e.Err
}

func (e *ConfigurationError) Error() string {
	return e.Message
}

func (e *ConfigurationError) Unwrap() error {
	return e.Err
}

func (e *RetryExhaustedError) Error() string {
	return fmt.Sprintf("all %d retry attempts failed: %v", e.Attempts, e.LastErr)
}

func (e *RetryExhaustedError) Unwrap() error {
	return e.LastErr
}

// Helper functions for creating errors

// NewAuthenticationError creates a new authentication error
func NewAuthenticationError(err error) error {
	return &AuthenticationError{Err: err}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(resource, id string, err error) error {
	return &NotFoundError{
		Resource: resource,
		ID:       id,
		Err:      err,
	}
}

// NewAPIError creates a new API error
func NewAPIError(statusCode int, message string, err error) error {
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
		Err:        err,
	}
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string, err error) error {
	return &ValidationError{
		Field:   field,
		Message: message,
		Err:     err,
	}
}

// NewNetworkError creates a new network error
func NewNetworkError(err error) error {
	return &NetworkError{Err: err}
}

// NewConfigurationError creates a new configuration error
func NewConfigurationError(message string, err error) error {
	return &ConfigurationError{
		Message: message,
		Err:     err,
	}
}

// NewRetryExhaustedError creates a new retry exhausted error
func NewRetryExhaustedError(attempts int, lastErr error) error {
	return &RetryExhaustedError{
		Attempts: attempts,
		LastErr:  lastErr,
	}
}

// IsAuthenticationError checks if an error is an authentication error
func IsAuthenticationError(err error) bool {
	var authErr *AuthenticationError
	return errors.As(err, &authErr)
}

// IsNotFoundError checks if an error is a not found error
func IsNotFoundError(err error) bool {
	var notFoundErr *NotFoundError
	return errors.As(err, &notFoundErr)
}

// IsAPIError checks if an error is an API error
func IsAPIError(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr)
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	var validationErr *ValidationError
	return errors.As(err, &validationErr)
}

// IsNetworkError checks if an error is a network error
func IsNetworkError(err error) bool {
	var networkErr *NetworkError
	return errors.As(err, &networkErr)
}

// IsConfigurationError checks if an error is a configuration error
func IsConfigurationError(err error) bool {
	var configErr *ConfigurationError
	return errors.As(err, &configErr)
}

// IsRetryExhaustedError checks if an error is a retry exhausted error
func IsRetryExhaustedError(err error) bool {
	var retryErr *RetryExhaustedError
	return errors.As(err, &retryErr)
}

// ErrorStyle defines the style for error messages
var ErrorStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("196")).
	Bold(true)

// WarningStyle defines the style for warning messages
var WarningStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("214"))

// SuggestionStyle defines the style for suggestion messages
var SuggestionStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("243")).
	Italic(true)

// FormatError formats an error for user display with actionable advice
func FormatError(err error) string {
	if err == nil {
		return ""
	}

	var message strings.Builder

	// Check for specific error types and provide actionable advice
	switch {
	case IsAuthenticationError(err):
		message.WriteString(ErrorStyle.Render("✗ Authentication failed"))
		message.WriteString("\n\n")
		message.WriteString("Your authentication token may have expired or been revoked.\n")
		message.WriteString("\n")
		message.WriteString(SuggestionStyle.Render("→ Run 'bc4 auth login' to refresh your credentials"))

	case IsNotFoundError(err):
		var notFoundErr *NotFoundError
		errors.As(err, &notFoundErr)
		message.WriteString(ErrorStyle.Render(fmt.Sprintf("✗ %s not found", notFoundErr.Resource)))
		message.WriteString("\n\n")
		if notFoundErr.ID != "" {
			message.WriteString(fmt.Sprintf("Could not find %s with ID '%s'.\n", strings.ToLower(notFoundErr.Resource), notFoundErr.ID))
			message.WriteString("\n")
			message.WriteString(SuggestionStyle.Render("→ Check the ID for typos"))
			message.WriteString("\n")
			message.WriteString(SuggestionStyle.Render(fmt.Sprintf("→ Run 'bc4 %s list' to see available items", strings.ToLower(notFoundErr.Resource))))
		} else {
			message.WriteString(fmt.Sprintf("No %s found matching your criteria.\n", strings.ToLower(notFoundErr.Resource)))
		}

	case IsAPIError(err):
		var apiErr *APIError
		errors.As(err, &apiErr)
		message.WriteString(ErrorStyle.Render("✗ API request failed"))
		message.WriteString("\n\n")

		switch apiErr.StatusCode {
		case http.StatusUnauthorized:
			message.WriteString("Authentication failed. Your token may be invalid or expired.\n")
			message.WriteString("\n")
			message.WriteString(SuggestionStyle.Render("→ Run 'bc4 auth login' to refresh your credentials"))
		case http.StatusForbidden:
			message.WriteString("You don't have permission to perform this action.\n")
			message.WriteString("\n")
			message.WriteString(SuggestionStyle.Render("→ Check that you have the necessary permissions in Basecamp"))
		case http.StatusNotFound:
			message.WriteString("The requested resource was not found.\n")
			message.WriteString("\n")
			message.WriteString(SuggestionStyle.Render("→ Verify the resource exists and you have access to it"))
		case http.StatusTooManyRequests:
			message.WriteString("Rate limit exceeded. Please wait before trying again.\n")
			message.WriteString("\n")
			message.WriteString(SuggestionStyle.Render("→ Wait a few moments and try again"))
		case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
			message.WriteString("Basecamp is experiencing issues. Please try again later.\n")
			message.WriteString("\n")
			message.WriteString(SuggestionStyle.Render("→ Check https://status.basecamp.com/ for service status"))
		default:
			message.WriteString(fmt.Sprintf("Request failed with status %d: %s\n", apiErr.StatusCode, apiErr.Message))
		}

	case IsValidationError(err):
		var validationErr *ValidationError
		errors.As(err, &validationErr)
		message.WriteString(ErrorStyle.Render("✗ Invalid input"))
		message.WriteString("\n\n")
		message.WriteString(validationErr.Error())
		message.WriteString("\n")
		message.WriteString("\n")
		message.WriteString(SuggestionStyle.Render("→ Check your input and try again"))

	case IsNetworkError(err):
		message.WriteString(ErrorStyle.Render("✗ Network connection failed"))
		message.WriteString("\n\n")
		message.WriteString("Unable to connect to Basecamp. This could be due to:\n")
		message.WriteString("  • No internet connection\n")
		message.WriteString("  • Firewall or proxy blocking the connection\n")
		message.WriteString("  • Basecamp service temporarily unavailable\n")
		message.WriteString("\n")
		message.WriteString(SuggestionStyle.Render("→ Check your internet connection and try again"))

	case IsConfigurationError(err):
		var configErr *ConfigurationError
		errors.As(err, &configErr)
		message.WriteString(ErrorStyle.Render("✗ Configuration error"))
		message.WriteString("\n\n")
		message.WriteString(configErr.Message)
		message.WriteString("\n")
		message.WriteString("\n")
		message.WriteString(SuggestionStyle.Render("→ Run 'bc4' to start the setup wizard"))

	default:
		// Handle OAuth-specific error messages
		errMsg := err.Error()
		if strings.Contains(errMsg, "OAuth credentials not configured") {
			message.WriteString(ErrorStyle.Render("✗ OAuth not configured"))
			message.WriteString("\n\n")
			message.WriteString("You need to set up OAuth credentials before using bc4.\n")
			message.WriteString("\n")
			message.WriteString(SuggestionStyle.Render("→ Run 'bc4' to start the setup wizard"))
		} else if strings.Contains(errMsg, "not authenticated") {
			message.WriteString(ErrorStyle.Render("✗ Not authenticated"))
			message.WriteString("\n\n")
			message.WriteString("You need to authenticate before using this command.\n")
			message.WriteString("\n")
			message.WriteString(SuggestionStyle.Render("→ Run 'bc4 auth login' to authenticate"))
		} else if strings.Contains(errMsg, "no account specified") {
			message.WriteString(ErrorStyle.Render("✗ No account selected"))
			message.WriteString("\n\n")
			message.WriteString("You need to select a default account or specify one with --account.\n")
			message.WriteString("\n")
			message.WriteString(SuggestionStyle.Render("→ Run 'bc4 account select' to choose a default account"))
			message.WriteString("\n")
			message.WriteString(SuggestionStyle.Render("→ Or use '--account <id>' to specify an account"))
		} else if strings.Contains(errMsg, "no project specified") {
			message.WriteString(ErrorStyle.Render("✗ No project selected"))
			message.WriteString("\n\n")
			message.WriteString("You need to select a default project or specify one with --project.\n")
			message.WriteString("\n")
			message.WriteString(SuggestionStyle.Render("→ Run 'bc4 project select' to choose a default project"))
			message.WriteString("\n")
			message.WriteString(SuggestionStyle.Render("→ Or use '--project <id>' to specify a project"))
		} else {
			// Generic error formatting
			message.WriteString(ErrorStyle.Render("✗ Error"))
			message.WriteString("\n\n")
			message.WriteString(err.Error())
		}
	}

	return message.String()
}
