package card

import (
	"fmt"

	"github.com/needmore/bc4/internal/factory"
	"github.com/spf13/cobra"
)

func newCreateCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Interactive card creation",
		Long:  `Create a new card using an interactive interface to select table, column, and add details.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement interactive card creation with Bubbletea
			return fmt.Errorf("interactive card creation not yet implemented")
		},
	}

	return cmd
}
