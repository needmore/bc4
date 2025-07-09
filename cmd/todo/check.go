package todo

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check [todo-id]",
		Short: "Mark a todo as complete",
		Long:  `Mark a todo as complete by its ID.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement marking todo as complete
			fmt.Println("Todo check not yet implemented")
			return nil
		},
	}

	return cmd
}
