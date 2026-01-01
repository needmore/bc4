package campfire

import (
	"github.com/needmore/bc4/internal/cmdutil"
	"github.com/needmore/bc4/internal/factory"
	"github.com/spf13/cobra"
)

// NewCampfireCmd creates the campfire command
func NewCampfireCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "campfire",
		Short: "Manage campfire chats",
		Long:  `Work with Basecamp campfires (chat rooms) - list, view messages, and post to campfires.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Enable suggestions for subcommand typos
	cmdutil.EnableSuggestions(cmd)

	// Add subcommands
	cmd.AddCommand(newListCmd(f))
	cmd.AddCommand(newSetCmd(f))
	cmd.AddCommand(newViewCmd(f))
	cmd.AddCommand(newPostCmd(f))

	return cmd
}
