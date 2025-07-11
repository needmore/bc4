package card

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newColumnMoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move [ID]",
		Short: "Reorder columns",
		Long:  `Move a column to a different position in the card table.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement column moving
			return fmt.Errorf("column moving not yet implemented")
		},
	}

	return cmd
}
