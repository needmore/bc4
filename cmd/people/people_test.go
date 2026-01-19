package people

import (
	"testing"

	"github.com/needmore/bc4/internal/factory"
	"github.com/stretchr/testify/assert"
)

func TestNewPeopleCmd(t *testing.T) {
	// Create factory
	f := factory.New()

	// Create people command
	cmd := NewPeopleCmd(f)

	// Test basic properties
	assert.Equal(t, "people", cmd.Use)
	assert.Contains(t, cmd.Aliases, "person")
	assert.Contains(t, cmd.Aliases, "users")
	assert.Contains(t, cmd.Aliases, "user")
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Test that subcommands are added
	subcommandNames := make(map[string]bool)
	for _, subcmd := range cmd.Commands() {
		subcommandNames[subcmd.Name()] = true
	}

	// Verify all expected subcommands are present
	expectedCommands := []string{
		"list",
		"view",
		"invite",
		"remove",
		"update",
		"ping",
	}

	for _, expected := range expectedCommands {
		assert.True(t, subcommandNames[expected], "Subcommand %s should be present", expected)
	}

	// Test that it has the correct number of subcommands
	assert.Len(t, cmd.Commands(), len(expectedCommands))
}

func TestPeopleCommand_Aliases(t *testing.T) {
	f := factory.New()
	cmd := NewPeopleCmd(f)

	// Test command aliases
	expectedAliases := []string{"person", "users", "user"}
	assert.Equal(t, expectedAliases, cmd.Aliases)
}

func TestPeopleCommand_SubcommandFactories(t *testing.T) {
	// Test that factory is properly passed to all subcommands
	f := factory.New()
	cmd := NewPeopleCmd(f)

	// Each subcommand should be properly initialized
	for _, subcmd := range cmd.Commands() {
		assert.NotNil(t, subcmd.RunE, "Subcommand %s should have RunE function", subcmd.Name())
		assert.NotEmpty(t, subcmd.Short, "Subcommand %s should have Short description", subcmd.Name())
	}
}
