package todo

import (
	"github.com/needmore/bc4/internal/factory"
	"github.com/spf13/cobra"
)

// NewTodoCmd creates the todo command
func NewTodoCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "todo",
		Short: "Work with Basecamp todos - list, view, create, and manage todos",
		Long: `Work with Basecamp todos and todo lists.

Basecamp projects can have multiple todo lists, each containing individual todos.
Use these commands to navigate and manage your tasks.`,
		Aliases: []string{"todos", "t"},
	}

	// Add subcommands
	cmd.AddCommand(newListsCmd(f))
	cmd.AddCommand(newListCmd(f))
	cmd.AddCommand(newSelectCmd(f))
	cmd.AddCommand(newSetCmd(f))
	cmd.AddCommand(newViewCmd(f))
	cmd.AddCommand(newAddCmd(f))
	cmd.AddCommand(newCheckCmd(f))
	cmd.AddCommand(newUncheckCmd(f))
	cmd.AddCommand(newCreateListCmd(f))

	return cmd
}
