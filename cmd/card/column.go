package card

import (
	"github.com/spf13/cobra"
)

// newColumnCmd creates the column management command
func newColumnCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "column",
		Short: "Manage card table columns",
		Long:  `Manage columns within card tables, including creating, editing, and reordering columns.`,
	}

	// Add subcommands
	cmd.AddCommand(newColumnListCmd())
	cmd.AddCommand(newColumnCreateCmd())
	cmd.AddCommand(newColumnEditCmd())
	cmd.AddCommand(newColumnMoveCmd())
	cmd.AddCommand(newColumnColorCmd())

	return cmd
}
