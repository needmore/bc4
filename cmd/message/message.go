package message

import (
	"github.com/needmore/bc4/internal/factory"
	"github.com/spf13/cobra"
)

// NewMessageCmd creates a new message command
func NewMessageCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "message",
		Short:   "Work with Basecamp messages",
		Long:    `Post and view messages in Basecamp projects.`,
		Aliases: []string{"messages", "msg"},
	}

	// Add subcommands
	cmd.AddCommand(newListCmd(f))
	cmd.AddCommand(newPostCmd(f))
	cmd.AddCommand(newViewCmd(f))
	cmd.AddCommand(newEditCmd(f))

	return cmd
}
