package comment

import (
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
		Example: `  bc4 comment list todo 123           # List comments on a todo
  bc4 comment create todo 123         # Add a comment to a todo
  bc4 comment view 456                # View a specific comment
  bc4 comment edit 456                # Edit a comment`,
	}

	// Add subcommands
	cmd.AddCommand(newListCmd(f))
	cmd.AddCommand(newViewCmd(f))
	cmd.AddCommand(newCreateCmd(f))
	cmd.AddCommand(newEditCmd(f))
	cmd.AddCommand(newDeleteCmd(f))

	return cmd
}
