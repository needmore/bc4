package schedule

import (
	"testing"

	"github.com/needmore/bc4/internal/factory"
	"github.com/stretchr/testify/assert"
)

func TestNewEntryListCmd(t *testing.T) {
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
			name:          "with upcoming flag",
			args:          []string{"--upcoming"},
			expectedError: false,
		},
		{
			name:          "with past flag",
			args:          []string{"--past"},
			expectedError: false,
		},
		{
			name:          "with range flag",
			args:          []string{"--range", "2025-01-01,2025-01-31"},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := factory.New()
			cmd := newEntryListCmd(f)
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

func TestEntryListCmdFlags(t *testing.T) {
	f := factory.New()
	cmd := newEntryListCmd(f)

	// Test upcoming flag exists
	upcomingFlag := cmd.Flags().Lookup("upcoming")
	assert.NotNil(t, upcomingFlag, "upcoming flag should exist")

	// Test past flag exists
	pastFlag := cmd.Flags().Lookup("past")
	assert.NotNil(t, pastFlag, "past flag should exist")

	// Test range flag exists
	rangeFlag := cmd.Flags().Lookup("range")
	assert.NotNil(t, rangeFlag, "range flag should exist")
}

func TestEntryListCmd_Properties(t *testing.T) {
	f := factory.New()
	cmd := newEntryListCmd(f)

	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "List schedule entries (calendar events)", cmd.Short)
	assert.Contains(t, cmd.Long, "By default, lists all entries")
}
