package card

import (
	"fmt"

	"github.com/needmore/bc4/internal/factory"
	"github.com/spf13/cobra"
)

func newColumnEditCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit [ID]",
		Short: "Edit column name/description",
		Long:  `Edit the name and description of a column.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement column editing
			return fmt.Errorf("column editing not yet implemented")
		},
	}

	return cmd
}
