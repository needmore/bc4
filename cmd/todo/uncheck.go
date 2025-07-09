package todo

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newUncheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uncheck [todo-id]",
		Short: "Mark a todo as incomplete",
		Long:  `Mark a todo as incomplete by its ID.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement marking todo as incomplete
			fmt.Println("Todo uncheck not yet implemented")
			return nil
		},
	}

	return cmd
}