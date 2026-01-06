package card

import (
	"github.com/needmore/bc4/internal/cmdutil"
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

	// Enable suggestions for subcommand typos
	cmdutil.EnableSuggestions(cmd)

	// Add subcommands
	cmd.AddCommand(newColumnListCmd(f))
	cmd.AddCommand(newColumnCreateCmd(f))
	cmd.AddCommand(newColumnEditCmd(f))
	cmd.AddCommand(newColumnMoveCmd(f))
	cmd.AddCommand(newColumnColorCmd(f))
	cmd.AddCommand(newColumnHoldCmd(f))
	cmd.AddCommand(newColumnUnholdCmd(f))

	return cmd
}
