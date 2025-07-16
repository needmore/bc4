package card

import (
	"testing"

	"github.com/needmore/bc4/internal/factory"
	"github.com/stretchr/testify/assert"
)

func TestNewCardCmd(t *testing.T) {
	f := factory.New()
	cmd := NewCardCmd(f)

	// Test basic command properties
	assert.Equal(t, "card", cmd.Use)
	assert.Equal(t, "Manage card tables and cards", cmd.Short)
	assert.Contains(t, cmd.Long, "Card tables are Basecamp's take on kanban")

	// Test that subcommands are added
	subcommands := []string{
		"list",
		"table",
		"view",
		"set",
		"add",
		"create",
		"edit",
		"move",
		"assign",
		"unassign",
		"archive",
		"column",
		"step",
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

// Table-driven tests for command parsing
func TestCardCmd_ParseFlags(t *testing.T) {
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
			name:          "invalid subcommand",
			args:          []string{"invalid"},
			expectedError: false, // Cobra doesn't error on unknown subcommands during parsing
		},
		{
			name:          "valid subcommand list",
			args:          []string{"list"},
			expectedError: false,
		},
		{
			name:          "empty args",
			args:          []string{},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := factory.New()
			cmd := NewCardCmd(f)

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

// Test that all subcommands have proper factory
func TestCardCmd_SubcommandsHaveFactory(t *testing.T) {
	f := factory.New()
	cmd := NewCardCmd(f)

	for _, subcmd := range cmd.Commands() {
		t.Run(subcmd.Name(), func(t *testing.T) {
			// Ensure subcommand is not nil
			assert.NotNil(t, subcmd)

			// Ensure subcommand has a Use field
			assert.NotEmpty(t, subcmd.Use)

			// Ensure subcommand has a Short description
			assert.NotEmpty(t, subcmd.Short)
		})
	}
}

// Test command hierarchy
func TestCardCmd_Hierarchy(t *testing.T) {
	f := factory.New()
	cmd := NewCardCmd(f)

	// Test that column command has subcommands
	columnCmd, _, err := cmd.Find([]string{"column"})
	assert.NoError(t, err)
	assert.NotNil(t, columnCmd)
	assert.True(t, len(columnCmd.Commands()) > 0, "column command should have subcommands")

	// Test that step command has subcommands
	stepCmd, _, err := cmd.Find([]string{"step"})
	assert.NoError(t, err)
	assert.NotNil(t, stepCmd)
	assert.True(t, len(stepCmd.Commands()) > 0, "step command should have subcommands")
}

// Benchmark command creation
func BenchmarkNewCardCmd(b *testing.B) {
	f := factory.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewCardCmd(f)
	}
}

// Test specific subcommand functionality
func TestCardCmd_CreateSubcommand(t *testing.T) {
	f := factory.New()
	cmd := NewCardCmd(f)

	// Find create subcommand
	createCmd, _, err := cmd.Find([]string{"create"})
	assert.NoError(t, err)
	assert.NotNil(t, createCmd)

	// Test create command flags
	assert.True(t, createCmd.Flags().HasFlags())

	// Check specific flags exist
	tableFlag := createCmd.Flags().Lookup("table")
	assert.NotNil(t, tableFlag)

	columnFlag := createCmd.Flags().Lookup("column")
	assert.NotNil(t, columnFlag)

	accountFlag := createCmd.Flags().Lookup("account")
	assert.NotNil(t, accountFlag)
	assert.Equal(t, "a", accountFlag.Shorthand)

	projectFlag := createCmd.Flags().Lookup("project")
	assert.NotNil(t, projectFlag)
	assert.Equal(t, "p", projectFlag.Shorthand)
}

// Test command examples
func TestCardCmd_Examples(t *testing.T) {
	f := factory.New()
	cmd := NewCardCmd(f)

	// Ensure examples are provided
	assert.NotEmpty(t, cmd.Example)

	// Check that examples contain common operations
	examples := []string{
		"bc4 card list",
		"bc4 card table",
		"bc4 card add",
		"bc4 card view",
		"bc4 card move",
	}

	for _, example := range examples {
		assert.Contains(t, cmd.Example, example)
	}
}

// Table-driven tests for subcommand validation
func TestCardCmd_SubcommandValidation(t *testing.T) {
	tests := []struct {
		name       string
		subcmd     string
		args       []string
		shouldFind bool
	}{
		{
			name:       "valid list command",
			subcmd:     "list",
			args:       []string{"list"},
			shouldFind: true,
		},
		{
			name:       "valid create command with flags",
			subcmd:     "create",
			args:       []string{"create", "--table", "123"},
			shouldFind: true,
		},
		{
			name:       "valid column list command",
			subcmd:     "column",
			args:       []string{"column", "list"},
			shouldFind: true,
		},
		{
			name:       "valid step add command",
			subcmd:     "step",
			args:       []string{"step", "add"},
			shouldFind: true,
		},
		{
			name:       "invalid subcommand",
			subcmd:     "invalid",
			args:       []string{"invalid"},
			shouldFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := factory.New()
			cmd := NewCardCmd(f)

			foundCmd, _, err := cmd.Find([]string{tt.subcmd})

			if tt.shouldFind {
				assert.NoError(t, err)
				assert.NotNil(t, foundCmd)
			} else {
				// Cobra returns the parent command when subcommand not found
				assert.Equal(t, cmd, foundCmd)
			}
		})
	}
}
