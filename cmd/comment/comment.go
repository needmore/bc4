package comment

import (
	"github.com/needmore/bc4/internal/factory"
	"github.com/spf13/cobra"
)

// NewCommentCmd creates a new comment command
func NewCommentCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comment",
		Short: "Work with Basecamp comments",
		Long:  `View, add, edit, and delete comments on Basecamp resources (todos, messages, documents, cards).`,
	}

	// Add subcommands
	cmd.AddCommand(newListCmd(f))
	cmd.AddCommand(newViewCmd(f))
	cmd.AddCommand(newCreateCmd(f))
	cmd.AddCommand(newEditCmd(f))
	cmd.AddCommand(newDeleteCmd(f))

	return cmd
}
