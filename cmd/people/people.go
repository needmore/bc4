package people

import (
	"github.com/needmore/bc4/internal/cmdutil"
	"github.com/needmore/bc4/internal/factory"
	"github.com/spf13/cobra"
)

// NewPeopleCmd creates the people command
func NewPeopleCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "people",
		Short: "Manage Basecamp people and project access",
		Long: `Work with Basecamp people - list team members, manage project access,
invite new people, and view who can be pinged.

People can be listed at the account level (all people) or filtered by project.
Project access can be managed by granting or revoking access to specific projects.`,
		Aliases: []string{"person", "users", "user"},
	}

	// Enable suggestions for subcommand typos
	cmdutil.EnableSuggestions(cmd)

	// Add subcommands
	cmd.AddCommand(newListCmd(f))
	cmd.AddCommand(newViewCmd(f))
	cmd.AddCommand(newInviteCmd(f))
	cmd.AddCommand(newRemoveCmd(f))
	cmd.AddCommand(newUpdateCmd(f))
	cmd.AddCommand(newPingCmd(f))

	return cmd
}
