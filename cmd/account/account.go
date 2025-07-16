package account

import (
	"github.com/needmore/bc4/internal/factory"
	"github.com/spf13/cobra"
)

// NewAccountCmd creates the account command
func NewAccountCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "account",
		Short:   "Manage Basecamp accounts",
		Long:    `Work with Basecamp accounts - list, select, and manage accounts.`,
		Aliases: []string{"a"},
	}

	// Add subcommands
	cmd.AddCommand(newListCmd(f))
	cmd.AddCommand(newSelectCmd(f))
	cmd.AddCommand(newSetCmd(f))
	cmd.AddCommand(newCurrentCmd(f))

	return cmd
}
