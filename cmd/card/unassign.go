package card

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newUnassignCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unassign [ID]",
		Short: "Remove assignees from card",
		Long:  `Remove one or more assignees from a card.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement card unassignment
			return fmt.Errorf("card unassignment not yet implemented")
		},
	}

	return cmd
}
