package card

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newMoveCmd() *cobra.Command {
	var columnName string

	cmd := &cobra.Command{
		Use:   "move [ID]",
		Short: "Move card between columns",
		Long:  `Move a card to a different column in the card table.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement card moving
			return fmt.Errorf("card moving not yet implemented")
		},
	}

	cmd.Flags().StringVar(&columnName, "column", "", "Target column name or ID")

	return cmd
}
