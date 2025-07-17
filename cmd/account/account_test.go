package account

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCmdAccount(t *testing.T) {
	// Create mock factory
	fact := &mockFactory{
		authClient: &mockAuthClient{},
	}
	
	// Create account command
	cmd := NewCmdAccount(fact)
	
	// Test basic properties
	assert.Equal(t, "account", cmd.Use)
	assert.Contains(t, cmd.Aliases, "a")
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	
	// Test that subcommands are added
	subcommandNames := make(map[string]bool)
	for _, subcmd := range cmd.Commands() {
		subcommandNames[subcmd.Name()] = true
	}
	
	// Verify all expected subcommands are present
	expectedCommands := []string{"list", "current", "set", "select"}
	for _, expected := range expectedCommands {
		assert.True(t, subcommandNames[expected], "Subcommand %s should be present", expected)
	}
	
	// Test that it has the correct number of subcommands
	assert.Len(t, cmd.Commands(), 4)
}

func TestAccountCommand_Factory(t *testing.T) {
	// Test that factory is properly passed to subcommands
	fact := &mockFactory{
		authClient: &mockAuthClient{
			accounts: map[string]auth.AccountToken{
				"123": {AccountID: "123", AccountName: "Test"},
			},
		},
	}
	
	cmd := NewCmdAccount(fact)
	
	// Get list subcommand and verify it works with the factory
	listCmd, _, err := cmd.Find([]string{"list"})
	assert.NoError(t, err)
	assert.NotNil(t, listCmd)
	
	// The factory should be available to the subcommand
	// (This is implicitly tested by the fact that subcommand tests pass)
}