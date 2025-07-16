package card

import (
	"fmt"

	"github.com/needmore/bc4/internal/factory"
	"github.com/spf13/cobra"
)

func newColumnColorCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "color [ID]",
		Short: "Set column color",
		Long:  `Set the color of a column. Available colors: white, red, orange, yellow, green, blue, aqua, purple, gray, pink, brown`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement column color setting
			return fmt.Errorf("column color setting not yet implemented")
		},
	}

	return cmd
}
