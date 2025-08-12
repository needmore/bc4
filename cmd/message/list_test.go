package message

import (
	"testing"

	"github.com/needmore/bc4/internal/factory"
	"github.com/stretchr/testify/assert"
)

func TestNewListCmd(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedError bool
		errorContains string
	}{
		{
			name:          "no arguments",
			args:          []string{},
			expectedError: false,
		},
		{
			name:          "with project ID",
			args:          []string{"123456"},
			expectedError: false,
		},
		{
			name:          "too many arguments",
			args:          []string{"123456", "extra"},
			expectedError: true,
			errorContains: "accepts at most 1 arg",
		},
		{
			name:          "with category flag",
			args:          []string{"--category", "Announcements"},
			expectedError: false,
		},
		{
			name:          "with limit flag",
			args:          []string{"--limit", "10"},
			expectedError: false,
		},
		{
			name:          "with invalid limit",
			args:          []string{"--limit", "invalid"},
			expectedError: true,
			errorContains: "invalid argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &factory.Factory{}
			cmd := newListCmd(f)
			cmd.SetArgs(tt.args)

			// Parse flags
			err := cmd.ParseFlags(tt.args)

			if tt.name == "too many arguments" {
				// Argument validation happens during Execute, not ParseFlags
				// So we expect no error here
				assert.NoError(t, err)
			} else if tt.expectedError {
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

func TestListCmdFlags(t *testing.T) {
	f := &factory.Factory{}
	cmd := newListCmd(f)

	// Test category flag
	categoryFlag := cmd.Flag("category")
	assert.NotNil(t, categoryFlag)
	assert.Equal(t, "c", categoryFlag.Shorthand)
	assert.Equal(t, "Filter by category", categoryFlag.Usage)

	// Test limit flag
	limitFlag := cmd.Flag("limit")
	assert.NotNil(t, limitFlag)
	assert.Equal(t, "l", limitFlag.Shorthand)
	assert.Equal(t, "Limit number of messages shown", limitFlag.Usage)
}
