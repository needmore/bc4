package project

import (
	"testing"

	"github.com/needmore/bc4/internal/factory"
	"github.com/stretchr/testify/assert"
)

func TestNewCmdProject(t *testing.T) {
	// Create factory
	f := factory.New()

	// Create project command
	cmd := NewProjectCmd(f)

	// Test basic properties
	assert.Equal(t, "project", cmd.Use)
	assert.Contains(t, cmd.Aliases, "p")
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
		"set",
		"select",
		"search",
	}

	for _, expected := range expectedCommands {
		assert.True(t, subcommandNames[expected], "Subcommand %s should be present", expected)
	}

	// Test that it has the correct number of subcommands
	assert.Len(t, cmd.Commands(), len(expectedCommands))
}

func TestProjectCommand_Examples(t *testing.T) {
	f := factory.New()
	cmd := NewProjectCmd(f)

	// Verify the command has a long description
	assert.NotEmpty(t, cmd.Long)
	assert.Contains(t, cmd.Long, "projects")
}

func TestProjectCommand_Aliases(t *testing.T) {
	f := factory.New()
	cmd := NewProjectCmd(f)

	// Test command aliases
	assert.Equal(t, []string{"p"}, cmd.Aliases)

	// The alias is in the command itself, not necessarily in the Long description
}

func TestProjectCommand_SubcommandFactories(t *testing.T) {
	// Test that factory is properly passed to all subcommands
	f := factory.New()
	cmd := NewProjectCmd(f)

	// Each subcommand should be properly initialized
	for _, subcmd := range cmd.Commands() {
		assert.NotNil(t, subcmd.RunE, "Subcommand %s should have RunE function", subcmd.Name())
		assert.NotEmpty(t, subcmd.Short, "Subcommand %s should have Short description", subcmd.Name())
	}
}
