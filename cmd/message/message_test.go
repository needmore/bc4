package message

import (
	"testing"

	"github.com/needmore/bc4/internal/factory"
	"github.com/stretchr/testify/assert"
)

func TestNewMessageCmd(t *testing.T) {
	f := &factory.Factory{}
	cmd := NewMessageCmd(f)

	assert.NotNil(t, cmd)
	assert.Equal(t, "message", cmd.Use)
	assert.Equal(t, "Work with Basecamp messages", cmd.Short)

	// Check aliases
	assert.Contains(t, cmd.Aliases, "messages")
	assert.Contains(t, cmd.Aliases, "msg")

	// Check subcommands
	subcommands := make(map[string]bool)
	for _, subcmd := range cmd.Commands() {
		subcommands[subcmd.Name()] = true
	}

	assert.True(t, subcommands["list"])
	assert.True(t, subcommands["post"])
	assert.True(t, subcommands["view"])
	assert.True(t, subcommands["edit"])
}

