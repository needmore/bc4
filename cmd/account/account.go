package account

import (
	"github.com/spf13/cobra"
)

// NewAccountCmd creates the account command
func NewAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "account",
		Short:   "Manage Basecamp accounts",
		Long:    `Work with Basecamp accounts - list, select, and manage accounts.`,
		Aliases: []string{"a"},
	}

	// Add subcommands
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newSelectCmd())
	cmd.AddCommand(newSetCmd())
	cmd.AddCommand(newCurrentCmd())

	return cmd
}
