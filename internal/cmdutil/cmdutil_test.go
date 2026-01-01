package cmdutil

import (
	"errors"
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestExitCodes(t *testing.T) {
	// Verify exit codes follow expected conventions
	assert.Equal(t, 0, ExitSuccess)
	assert.Equal(t, 1, ExitError)
	assert.Equal(t, 2, ExitUsageError)
	assert.Equal(t, 4, ExitAuthError)
	assert.Equal(t, 5, ExitNotFound)
	assert.Equal(t, 130, ExitCancelled)
}

func TestEnableSuggestions(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	assert.Equal(t, 0, cmd.SuggestionsMinimumDistance)

	EnableSuggestions(cmd)
	assert.Equal(t, 2, cmd.SuggestionsMinimumDistance)
}

func TestDisableAuthCheck(t *testing.T) {
	t.Run("sets annotation on command with nil annotations", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		assert.Nil(t, cmd.Annotations)

		DisableAuthCheck(cmd)
		assert.NotNil(t, cmd.Annotations)
		assert.Equal(t, "true", cmd.Annotations["skipAuth"])
	})

	t.Run("sets annotation on command with existing annotations", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.Annotations = map[string]string{"existing": "value"}

		DisableAuthCheck(cmd)
		assert.Equal(t, "true", cmd.Annotations["skipAuth"])
		assert.Equal(t, "value", cmd.Annotations["existing"])
	})
}

func TestRequiresAuth(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]string
		expected    bool
	}{
		{
			name:        "nil annotations requires auth",
			annotations: nil,
			expected:    true,
		},
		{
			name:        "empty annotations requires auth",
			annotations: map[string]string{},
			expected:    true,
		},
		{
			name:        "skipAuth false requires auth",
			annotations: map[string]string{"skipAuth": "false"},
			expected:    true,
		},
		{
			name:        "skipAuth true does not require auth",
			annotations: map[string]string{"skipAuth": "true"},
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			cmd.Annotations = tt.annotations
			assert.Equal(t, tt.expected, RequiresAuth(cmd))
		})
	}
}

func TestExactArgs(t *testing.T) {
	cmd := &cobra.Command{Use: "test <id>"}
	validator := ExactArgs(1, "id")

	tests := []struct {
		name      string
		args      []string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "exact number of args",
			args:      []string{"123"},
			expectErr: false,
		},
		{
			name:      "too few args",
			args:      []string{},
			expectErr: true,
			errMsg:    "missing required argument: <id>",
		},
		{
			name:      "too many args",
			args:      []string{"123", "456"},
			expectErr: true,
			errMsg:    "too many arguments (expected 1, got 2)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(cmd, tt.args)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.True(t, IsUsageError(err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMinArgs(t *testing.T) {
	cmd := &cobra.Command{Use: "test <files...>"}
	validator := MinArgs(1, "file")

	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "minimum args",
			args:      []string{"file1"},
			expectErr: false,
		},
		{
			name:      "more than minimum",
			args:      []string{"file1", "file2", "file3"},
			expectErr: false,
		},
		{
			name:      "no args",
			args:      []string{},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(cmd, tt.args)
			if tt.expectErr {
				assert.Error(t, err)
				assert.True(t, IsUsageError(err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRangeArgs(t *testing.T) {
	cmd := &cobra.Command{Use: "test <source> [dest]"}
	validator := RangeArgs(1, 2, "source")

	tests := []struct {
		name      string
		args      []string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "minimum args",
			args:      []string{"src"},
			expectErr: false,
		},
		{
			name:      "maximum args",
			args:      []string{"src", "dest"},
			expectErr: false,
		},
		{
			name:      "no args",
			args:      []string{},
			expectErr: true,
			errMsg:    "missing required argument",
		},
		{
			name:      "too many args",
			args:      []string{"a", "b", "c"},
			expectErr: true,
			errMsg:    "too many arguments (expected at most 2, got 3)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(cmd, tt.args)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.True(t, IsUsageError(err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUsageError(t *testing.T) {
	cmd := &cobra.Command{
		Use:     "test <id>",
		Example: "bc4 test 12345",
	}

	err := &UsageError{
		Message: "missing required argument",
		Cmd:     cmd,
	}

	t.Run("error message includes usage", func(t *testing.T) {
		msg := err.Error()
		assert.Contains(t, msg, "missing required argument")
		assert.Contains(t, msg, "Usage:")
		assert.Contains(t, msg, "test <id>")
	})

	t.Run("error message includes example", func(t *testing.T) {
		msg := err.Error()
		assert.Contains(t, msg, "Examples:")
		assert.Contains(t, msg, "bc4 test 12345")
	})

	t.Run("error message without example", func(t *testing.T) {
		cmdNoExample := &cobra.Command{Use: "test"}
		errNoExample := &UsageError{
			Message: "error",
			Cmd:     cmdNoExample,
		}
		msg := errNoExample.Error()
		assert.NotContains(t, msg, "Examples:")
	})
}

func TestIsUsageError(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "usage error",
			err:      &UsageError{Message: "test", Cmd: cmd},
			expected: true,
		},
		{
			name:     "wrapped usage error",
			err:      fmt.Errorf("wrapped: %w", &UsageError{Message: "test", Cmd: cmd}),
			expected: true,
		},
		{
			name:     "regular error",
			err:      errors.New("regular error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsUsageError(tt.err))
		})
	}
}

func TestSilentError(t *testing.T) {
	innerErr := errors.New("inner error")
	silentErr := NewSilentError(innerErr)

	t.Run("error message", func(t *testing.T) {
		assert.Equal(t, "inner error", silentErr.Error())
	})

	t.Run("unwrap", func(t *testing.T) {
		unwrapped := errors.Unwrap(silentErr)
		assert.Equal(t, innerErr, unwrapped)
	})
}

func TestIsSilentError(t *testing.T) {
	innerErr := errors.New("inner error")

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "silent error",
			err:      NewSilentError(innerErr),
			expected: true,
		},
		{
			name:     "wrapped silent error",
			err:      fmt.Errorf("wrapped: %w", NewSilentError(innerErr)),
			expected: true,
		},
		{
			name:     "regular error",
			err:      innerErr,
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsSilentError(tt.err))
		})
	}
}

func TestUnwrapSilent(t *testing.T) {
	innerErr := errors.New("inner error")

	tests := []struct {
		name     string
		err      error
		expected error
	}{
		{
			name:     "silent error",
			err:      NewSilentError(innerErr),
			expected: innerErr,
		},
		{
			name:     "regular error",
			err:      innerErr,
			expected: innerErr,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, UnwrapSilent(tt.err))
		})
	}
}
