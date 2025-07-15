package card

import (
	"fmt"

	"github.com/needmore/bc4/internal/factory"
	"github.com/spf13/cobra"
)

func newColumnListCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List columns in current table",
		Long:  `List all columns in the current card table.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement column listing
			return fmt.Errorf("column listing not yet implemented")
		},
	}

	return cmd
}
