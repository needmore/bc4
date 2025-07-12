package card

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newEditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit [ID]",
		Short: "Edit card title/content",
		Long:  `Edit the title and content of an existing card.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement card editing
			return fmt.Errorf("card editing not yet implemented")
		},
	}

	return cmd
}
