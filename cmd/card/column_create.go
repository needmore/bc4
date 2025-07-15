package card

import (
	"fmt"

	"github.com/needmore/bc4/internal/factory"
	"github.com/spf13/cobra"
)

func newColumnCreateCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create new column",
		Long:  `Create a new column in the current card table.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement column creation
			return fmt.Errorf("column creation not yet implemented")
		},
	}

	return cmd
}
