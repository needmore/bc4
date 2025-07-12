package campfire

import (
	"github.com/spf13/cobra"
)

// NewCampfireCmd creates the campfire command
func NewCampfireCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "campfire",
		Short: "Manage campfire chats",
		Long:  `Work with Basecamp campfires (chat rooms) - list, view messages, and post to campfires.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Add subcommands
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newSetCmd())
	cmd.AddCommand(newViewCmd())
	cmd.AddCommand(newPostCmd())

	return cmd
}
