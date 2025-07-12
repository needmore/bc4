package card

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newAssignCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assign [ID]",
		Short: "Assign people to card",
		Long:  `Assign one or more people to a card.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement card assignment
			return fmt.Errorf("card assignment not yet implemented")
		},
	}

	return cmd
}
