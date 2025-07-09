package todo

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newViewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view [list-id|name]",
		Short: "View todos in a specific list",
		Long:  `View all todos in a specific todo list. Can specify by ID or partial name match.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement viewing todos in a list
			fmt.Println("Todo view not yet implemented")
			return nil
		},
	}

	return cmd
}