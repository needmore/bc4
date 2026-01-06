package comment

import (
	"github.com/needmore/bc4/internal/cmdutil"
	"github.com/needmore/bc4/internal/factory"
	"github.com/spf13/cobra"
)

// NewCommentCmd creates a new comment command
func NewCommentCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "comment",
		Short:   "Work with Basecamp comments",
		Aliases: []string{"comments"},
		Long: `View, add, edit, and delete comments on Basecamp resources.

Comments can be added to various Basecamp resources including todos, messages,
documents, and cards. Use these commands to view and manage conversations
on your team's work.`,
		Example: `  bc4 comment list 12345              # List comments on a recording (todo, message, etc.)
  bc4 comment create 12345            # Add a comment to a recording
  bc4 comment view 67890              # View a specific comment
  bc4 comment edit 67890              # Edit a comment`,
	}

	// Enable suggestions for subcommand typos
	cmdutil.EnableSuggestions(cmd)

	// Add subcommands
	cmd.AddCommand(newListCmd(f))
	cmd.AddCommand(newViewCmd(f))
	cmd.AddCommand(newCreateCmd(f))
	cmd.AddCommand(newEditCmd(f))
	cmd.AddCommand(newAttachCmd(f))
	cmd.AddCommand(newDeleteCmd(f))

	return cmd
}
