package todo

import (
	"fmt"

	"github.com/needmore/bc4/internal/factory"
	"github.com/spf13/cobra"
)

func newSelectCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "select",
		Short: "Select default todo list",
		Long:  `Interactively select a default todo list for the current project.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement interactive todo list selection
			fmt.Println("Todo list selection not yet implemented")
			return nil
		},
	}

	return cmd
}
