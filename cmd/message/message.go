package message

import (
	"github.com/needmore/bc4/internal/cmdutil"
	"github.com/needmore/bc4/internal/factory"
	"github.com/spf13/cobra"
)

// NewMessageCmd creates a new message command
func NewMessageCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "message",
		Short:   "Work with Basecamp messages",
		Aliases: []string{"messages", "msg"},
		Long: `Post and view messages in Basecamp projects.

Messages are the primary way to share announcements, updates, and discussions
with your team in Basecamp. They support rich formatting and can include
attachments. Team members can comment on messages to continue the conversation.`,
		Example: `  bc4 message list                    # List recent messages
  bc4 message post "Team Update"      # Post a new message
  bc4 message view 123                # View message details
  bc4 message edit 123                # Edit an existing message
  bc4 message pin 123                 # Pin a message to the top
  bc4 message unpin 123               # Unpin a message`,
	}

	// Enable suggestions for subcommand typos
	cmdutil.EnableSuggestions(cmd)

	// Add subcommands
	cmd.AddCommand(newListCmd(f))
	cmd.AddCommand(newPostCmd(f))
	cmd.AddCommand(newViewCmd(f))
	cmd.AddCommand(newEditCmd(f))
	cmd.AddCommand(newPinCmd(f))
	cmd.AddCommand(newUnpinCmd(f))

	return cmd
}
