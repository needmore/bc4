package todo

import (
	"fmt"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/spf13/cobra"
)

type createListOptions struct {
	description string
}

func newCreateListCmd(f *factory.Factory) *cobra.Command {
	opts := &createListOptions{}

	cmd := &cobra.Command{
		Use:   "create-list [<name>]",
		Short: "Create a new todo list",
		Long: `Create a new todo list in the current project.

If no name is provided, you'll be prompted to enter one interactively.`,
		Example: `  # Create a new todo list
  bc4 todo create-list "Sprint 1 Tasks"

  # Create a todo list with description
  bc4 todo create-list "Bug Fixes" --description "Critical bugs to fix before release"`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreateList(f, opts, args)
		},
	}

	cmd.Flags().StringVarP(&opts.description, "description", "d", "", "Description for the todo list")

	return cmd
}

func runCreateList(f *factory.Factory, opts *createListOptions, args []string) error {
	// Get name from args or prompt
	var name string
	if len(args) > 0 {
		name = args[0]
	} else {
		// TODO: Add interactive prompt using Bubbletea
		return fmt.Errorf("interactive mode not yet implemented. Please provide a name as an argument")
	}

	// Get API client from factory
	client, err := f.ApiClient()
	if err != nil {
		return err
	}
	todoOps := client.Todos()

	// Get resolved project ID
	projectID, err := f.ProjectID()
	if err != nil {
		return err
	}

	// Get the todo set for the project
	todoSet, err := todoOps.GetProjectTodoSet(f.Context(), projectID)
	if err != nil {
		return fmt.Errorf("failed to get todo set: %w", err)
	}

	// Create the todo list
	req := api.TodoListCreateRequest{
		Name:        name,
		Description: opts.description,
	}

	todoList, err := todoOps.CreateTodoList(f.Context(), projectID, todoSet.ID, req)
	if err != nil {
		return fmt.Errorf("failed to create todo list: %w", err)
	}

	// GitHub CLI style: minimal output with created ID
	fmt.Printf("Created todo list #%d\n", todoList.ID)

	// Suggest setting as default
	fmt.Printf("\nTo set as default: bc4 todo set %d\n", todoList.ID)

	return nil
}
