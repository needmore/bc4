package card

import (
	"github.com/needmore/bc4/internal/factory"
	"github.com/spf13/cobra"
)

// newColumnCmd creates the column management command
func newColumnCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "column",
		Short: "Manage card table columns",
		Long:  `Manage columns within card tables, including creating, editing, and reordering columns.`,
	}

	// Add subcommands
	cmd.AddCommand(newColumnListCmd(f))
	cmd.AddCommand(newColumnCreateCmd(f))
	cmd.AddCommand(newColumnEditCmd(f))
	cmd.AddCommand(newColumnMoveCmd(f))
	cmd.AddCommand(newColumnColorCmd(f))

	return cmd
}
