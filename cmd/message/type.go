package message

import (
	"github.com/needmore/bc4/internal/cmdutil"
	"github.com/needmore/bc4/internal/factory"
	"github.com/spf13/cobra"
)

// newTypeCmd creates the message type management command
func newTypeCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "type",
		Short:   "Manage message categories",
		Aliases: []string{"category", "categories"},
		Long: `Manage message board categories for better organization.

Message categories allow you to organize messages into different types
(e.g., announcements, updates, questions). This helps teams find and
filter messages more effectively.`,
		Example: `  bc4 message type list               # List all categories
  bc4 message type create "Announcements"
  bc4 message type edit 123 --name "Updates"
  bc4 message type delete 123`,
	}

	// Enable suggestions for subcommand typos
	cmdutil.EnableSuggestions(cmd)

	// Add subcommands
	cmd.AddCommand(newTypeListCmd(f))
	cmd.AddCommand(newTypeCreateCmd(f))
	cmd.AddCommand(newTypeEditCmd(f))
	cmd.AddCommand(newTypeDeleteCmd(f))

	return cmd
}
