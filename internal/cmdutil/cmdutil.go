// Package cmdutil provides CLI command utilities for consistent behavior
// across all bc4 commands, including suggestions, argument validation,
// and exit code management.
package cmdutil

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// Exit codes following GitHub CLI conventions
const (
	ExitSuccess       = 0 // Successful execution
	ExitError         = 1 // General error
	ExitUsageError    = 2 // Invalid command usage
	ExitAuthError     = 4 // Authentication failure
	ExitNotFound      = 5 // Resource not found
	ExitCanceled      = 130 // User canceled (Ctrl+C)
)

// EnableSuggestions configures a command to suggest similar commands on typos.
// This should be called on all parent commands that have subcommands.
func EnableSuggestions(cmd *cobra.Command) {
	cmd.SuggestionsMinimumDistance = 2
}

// DisableAuthCheck marks a command as not requiring authentication.
// It sets an annotation that can be checked with RequiresAuth.
func DisableAuthCheck(cmd *cobra.Command) {
	if cmd.Annotations == nil {
		cmd.Annotations = make(map[string]string)
	}
	cmd.Annotations["skipAuth"] = "true"
}

// RequiresAuth returns true if the command requires authentication
func RequiresAuth(cmd *cobra.Command) bool {
	if cmd.Annotations == nil {
		return true
	}
	return cmd.Annotations["skipAuth"] != "true"
}

// ExactArgs returns an argument validator that requires exactly n arguments
// with a helpful error message including the command usage.
func ExactArgs(n int, argName string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < n {
			return &UsageError{
				Message: fmt.Sprintf("missing required argument: <%s>", argName),
				Cmd:     cmd,
			}
		}
		if len(args) > n {
			return &UsageError{
				Message: fmt.Sprintf("too many arguments (expected %d, got %d)", n, len(args)),
				Cmd:     cmd,
			}
		}
		return nil
	}
}

// MinArgs returns an argument validator that requires at least n arguments
// with a helpful error message.
func MinArgs(n int, argName string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < n {
			return &UsageError{
				Message: fmt.Sprintf("missing required argument: <%s>", argName),
				Cmd:     cmd,
			}
		}
		return nil
	}
}

// RangeArgs returns an argument validator that requires between min and max arguments.
func RangeArgs(min, max int, argName string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < min {
			return &UsageError{
				Message: fmt.Sprintf("missing required argument: <%s>", argName),
				Cmd:     cmd,
			}
		}
		if len(args) > max {
			return &UsageError{
				Message: fmt.Sprintf("too many arguments (expected at most %d, got %d)", max, len(args)),
				Cmd:     cmd,
			}
		}
		return nil
	}
}

// UsageError represents an error in command usage that should show help context
type UsageError struct {
	Message string
	Cmd     *cobra.Command
}

func (e *UsageError) Error() string {
	var sb strings.Builder
	sb.WriteString(e.Message)
	sb.WriteString("\n\n")
	sb.WriteString("Usage:\n  ")
	sb.WriteString(e.Cmd.UseLine())

	// Add example if available
	if e.Cmd.Example != "" {
		sb.WriteString("\n\nExamples:\n")
		sb.WriteString(e.Cmd.Example)
	}

	return sb.String()
}

// IsUsageError checks if an error is a usage error
func IsUsageError(err error) bool {
	var usageErr *UsageError
	return errors.As(err, &usageErr)
}

// SilentError wraps an error that has already been displayed to the user.
// This allows returning a proper exit code without double-printing the error.
type SilentError struct {
	Err error
}

func (e *SilentError) Error() string {
	return e.Err.Error()
}

func (e *SilentError) Unwrap() error {
	return e.Err
}

// NewSilentError creates a silent error wrapper
func NewSilentError(err error) error {
	return &SilentError{Err: err}
}

// IsSilentError checks if an error is a silent error
func IsSilentError(err error) bool {
	var silentErr *SilentError
	return errors.As(err, &silentErr)
}

// UnwrapSilent returns the underlying error from a SilentError
func UnwrapSilent(err error) error {
	var silentErr *SilentError
	if errors.As(err, &silentErr) {
		return silentErr.Err
	}
	return err
}
