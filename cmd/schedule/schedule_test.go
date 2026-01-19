package schedule

import (
	"testing"

	"github.com/needmore/bc4/internal/factory"
	"github.com/stretchr/testify/assert"
)

func TestNewScheduleCmd(t *testing.T) {
	f := factory.New()
	cmd := NewScheduleCmd(f)

	// Test basic command properties
	assert.Equal(t, "schedule", cmd.Use)
	assert.Equal(t, "Work with Basecamp schedules - calendar events and entries", cmd.Short)
	assert.Contains(t, cmd.Long, "Each Basecamp project can have a schedule")

	// Check aliases
	assert.Contains(t, cmd.Aliases, "schedules")
	assert.Contains(t, cmd.Aliases, "cal")
	assert.Contains(t, cmd.Aliases, "calendar")

	// Test that subcommands are added
	subcommands := []string{
		"list",
		"view",
		"entry",
	}

	for _, subcmd := range subcommands {
		t.Run("has_"+subcmd+"_subcommand", func(t *testing.T) {
			found := false
			for _, c := range cmd.Commands() {
				if c.Name() == subcmd {
					found = true
					break
				}
			}
			assert.True(t, found, "Expected to find subcommand: %s", subcmd)
		})
	}
}

func TestNewEntryCmd(t *testing.T) {
	f := factory.New()
	cmd := newEntryCmd(f)

	// Test basic command properties
	assert.Equal(t, "entry", cmd.Use)
	assert.Equal(t, "Manage schedule entries (calendar events)", cmd.Short)
	assert.Contains(t, cmd.Long, "Schedule entries can be")

	// Check aliases
	assert.Contains(t, cmd.Aliases, "entries")
	assert.Contains(t, cmd.Aliases, "event")
	assert.Contains(t, cmd.Aliases, "events")

	// Test that subcommands are added
	subcommands := []string{
		"list",
		"view",
		"create",
		"edit",
		"delete",
	}

	for _, subcmd := range subcommands {
		t.Run("has_"+subcmd+"_subcommand", func(t *testing.T) {
			found := false
			for _, c := range cmd.Commands() {
				if c.Name() == subcmd {
					found = true
					break
				}
			}
			assert.True(t, found, "Expected to find subcommand: %s", subcmd)
		})
	}
}

func TestScheduleCmd_ParseFlags(t *testing.T) {
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
			name:          "help flag",
			args:          []string{"--help"},
			expectedError: true,
			errorContains: "help requested",
		},
		{
			name:          "valid subcommand list",
			args:          []string{"list"},
			expectedError: false,
		},
		{
			name:          "valid subcommand entry",
			args:          []string{"entry"},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := factory.New()
			cmd := NewScheduleCmd(f)

			// Set args
			cmd.SetArgs(tt.args)

			// Parse flags only (don't execute)
			err := cmd.ParseFlags(tt.args)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestScheduleCmd_SubcommandsHaveFactory(t *testing.T) {
	f := factory.New()
	cmd := NewScheduleCmd(f)

	// Verify each subcommand can be created and has appropriate structure
	subcommands := []string{"list", "view", "entry"}
	for _, subcmd := range subcommands {
		t.Run(subcmd, func(t *testing.T) {
			found := false
			for _, c := range cmd.Commands() {
				if c.Name() == subcmd {
					found = true
					assert.NotNil(t, c, "Subcommand %s should not be nil", subcmd)
					break
				}
			}
			assert.True(t, found, "Expected to find subcommand: %s", subcmd)
		})
	}
}
