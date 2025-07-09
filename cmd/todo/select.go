package todo

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newSelectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "select",
		Short: "Select default todo list",
		Long:  `Interactively select a default todo list for the current project.`,
		Aliases: []string{"set-default", "default"},
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement interactive todo list selection
			fmt.Println("Todo list selection not yet implemented")
			return nil
		},
	}

	return cmd
}