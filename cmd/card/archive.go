package card

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newArchiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive [ID]",
		Short: "Archive a card",
		Long:  `Archive a card, removing it from the active board.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement card archiving
			return fmt.Errorf("card archiving not yet implemented")
		},
	}

	return cmd
}
