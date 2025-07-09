package todo

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newAddCmd() *cobra.Command {
	var listID string

	cmd := &cobra.Command{
		Use:   "add [title]",
		Short: "Add a new todo",
		Long:  `Add a new todo to a todo list. If no list is specified, uses the default list.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement adding a todo
			fmt.Println("Todo add not yet implemented")
			return nil
		},
	}

	cmd.Flags().StringVarP(&listID, "list", "l", "", "Todo list ID or name")

	return cmd
}