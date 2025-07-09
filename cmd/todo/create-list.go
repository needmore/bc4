package todo

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newCreateListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-list [name]",
		Short: "Create a new todo list",
		Long:  `Create a new todo list in the current project.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement creating a new todo list
			fmt.Println("Todo create-list not yet implemented")
			return nil
		},
	}

	return cmd
}