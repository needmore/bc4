package todo

import (
	"github.com/spf13/cobra"
)

// NewTodoCmd creates the todo command
func NewTodoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "todo",
		Short: "Work with Basecamp todos - list, view, create, and manage todos",
		Long: `Work with Basecamp todos and todo lists.

Basecamp projects can have multiple todo lists, each containing individual todos.
Use these commands to navigate and manage your tasks.`,
		Aliases: []string{"todos", "t"},
	}

	// Add subcommands
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newSelectCmd())
	cmd.AddCommand(newSetCmd())
	cmd.AddCommand(newViewCmd())
	cmd.AddCommand(newAddCmd())
	cmd.AddCommand(newCheckCmd())
	cmd.AddCommand(newUncheckCmd())
	cmd.AddCommand(newCreateListCmd())

	return cmd
}
