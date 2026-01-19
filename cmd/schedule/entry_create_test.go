package schedule

import (
	"testing"
	"time"

	"github.com/needmore/bc4/internal/factory"
	"github.com/stretchr/testify/assert"
)

func TestNewEntryCreateCmd(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedError bool
		errorContains string
	}{
		{
			name:          "no arguments",
			args:          []string{},
			expectedError: false, // ParseFlags won't error on missing args
		},
		{
			name:          "with title and required flags",
			args:          []string{"Meeting", "--starts-at", "2025-01-20T14:00:00"},
			expectedError: false,
		},
		{
			name:          "with all-day flag",
			args:          []string{"Holiday", "--starts-at", "2025-01-20", "--all-day"},
			expectedError: false,
		},
		{
			name:          "with participants",
			args:          []string{"Meeting", "--starts-at", "2025-01-20T14:00:00", "--participant", "user@example.com"},
			expectedError: false,
		},
		{
			name:          "with notify flag",
			args:          []string{"Meeting", "--starts-at", "2025-01-20T14:00:00", "--notify"},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := factory.New()
			cmd := newEntryCreateCmd(f)
			cmd.SetArgs(tt.args)

			// Parse flags
			err := cmd.ParseFlags(tt.args)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" && err != nil {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEntryCreateCmdFlags(t *testing.T) {
	f := factory.New()
	cmd := newEntryCreateCmd(f)

	// Test required and optional flags exist
	flags := []string{
		"description",
		"starts-at",
		"ends-at",
		"all-day",
		"participant",
		"notify",
	}

	for _, flagName := range flags {
		t.Run("has_"+flagName+"_flag", func(t *testing.T) {
			flag := cmd.Flags().Lookup(flagName)
			assert.NotNil(t, flag, "%s flag should exist", flagName)
		})
	}
}

func TestEntryCreateCmd_Properties(t *testing.T) {
	f := factory.New()
	cmd := newEntryCreateCmd(f)

	assert.Equal(t, "create <title>", cmd.Use)
	assert.Equal(t, "Create a new schedule entry (calendar event)", cmd.Short)
	assert.Contains(t, cmd.Long, "Events can be")
}

func TestParseDateTime(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		allDay       bool
		expectError  bool
		validateFunc func(t *testing.T, result string)
	}{
		{
			name:        "ISO 8601 with time",
			input:       "2025-01-20T14:30:00",
			allDay:      false,
			expectError: false,
			validateFunc: func(t *testing.T, result string) {
				assert.Contains(t, result, "2025-01-20")
				assert.Contains(t, result, "14:30")
			},
		},
		{
			name:        "ISO 8601 date only",
			input:       "2025-01-20",
			allDay:      false,
			expectError: false,
			validateFunc: func(t *testing.T, result string) {
				assert.Contains(t, result, "2025-01-20")
			},
		},
		{
			name:        "ISO 8601 date only as all-day",
			input:       "2025-01-20",
			allDay:      true,
			expectError: false,
			validateFunc: func(t *testing.T, result string) {
				assert.Equal(t, "2025-01-20", result)
			},
		},
		{
			name:        "relative date: today",
			input:       "today",
			allDay:      true,
			expectError: false,
			validateFunc: func(t *testing.T, result string) {
				today := time.Now().Format("2006-01-02")
				assert.Equal(t, today, result)
			},
		},
		{
			name:        "relative date: tomorrow",
			input:       "tomorrow",
			allDay:      true,
			expectError: false,
			validateFunc: func(t *testing.T, result string) {
				tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
				assert.Equal(t, tomorrow, result)
			},
		},
		{
			name:        "relative date: next-monday",
			input:       "next-monday",
			allDay:      false,
			expectError: false,
			validateFunc: func(t *testing.T, result string) {
				assert.NotEmpty(t, result)
			},
		},
		{
			name:        "invalid format",
			input:       "not-a-date",
			allDay:      false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDateTime(tt.input, tt.allDay)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validateFunc != nil {
					tt.validateFunc(t, result)
				}
			}
		})
	}
}
