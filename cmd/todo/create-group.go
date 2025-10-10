package todo

import (
	"fmt"
	"strings"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
)

type createGroupOptions struct {
	list string
}

func newCreateGroupCmd(f *factory.Factory) *cobra.Command {
	opts := &createGroupOptions{}

	cmd := &cobra.Command{
		Use:   "create-group [<name>]",
		Short: "Create a new group within a todo list",
		Long: `Create a new group (section) within a todo list to organize todos.

If no name is provided, you'll be prompted to enter one interactively.
Groups allow you to organize todos into sections within a list.`,
		Example: `  # Create a new group in the default todo list
  bc4 todo create-group "In Progress"

  # Create a group in a specific list
  bc4 todo create-group "Completed" --list "Sprint 1 Tasks"

  # Create a group using list ID
  bc4 todo create-group "Backlog" --list 12345`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreateGroup(f, opts, args)
		},
	}

	cmd.Flags().StringVarP(&opts.list, "list", "l", "", "Todo list ID, name, or URL (defaults to selected list)")

	return cmd
}

func runCreateGroup(f *factory.Factory, opts *createGroupOptions, args []string) error {
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

	// Get config for default lookups
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	// Get resolved account ID
	resolvedAccountID, err := f.AccountID()
	if err != nil {
		return err
	}

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

	// Resolve todo list ID
	var todoListID int64
	if opts.list != "" {
		// Check if it's a URL
		if parser.IsBasecampURL(opts.list) {
			parsed, err := parser.ParseBasecampURL(opts.list)
			if err != nil {
				return fmt.Errorf("invalid Basecamp URL: %w", err)
			}
			if parsed.ResourceType != parser.ResourceTypeTodoList {
				return fmt.Errorf("URL is not a todo list URL: %s", opts.list)
			}
			todoListID = parsed.ResourceID
		} else {
			// User specified a list - try to find it
			todoLists, err := todoOps.GetTodoLists(f.Context(), projectID, todoSet.ID)
			if err != nil {
				return fmt.Errorf("failed to fetch todo lists: %w", err)
			}

			// Try to match by ID or name
			for _, list := range todoLists {
				if fmt.Sprintf("%d", list.ID) == opts.list ||
					strings.EqualFold(list.Title, opts.list) ||
					strings.EqualFold(list.Name, opts.list) {
					todoListID = list.ID
					break
				}
			}

			if todoListID == 0 {
				return fmt.Errorf("todo list not found: %s", opts.list)
			}
		}
	} else {
		// Use default todo list from config
		defaultTodoListID := ""
		if cfg.Accounts != nil {
			if acc, ok := cfg.Accounts[resolvedAccountID]; ok {
				if acc.ProjectDefaults != nil {
					if projDefaults, ok := acc.ProjectDefaults[projectID]; ok {
						defaultTodoListID = projDefaults.DefaultTodoList
					}
				}
			}
		}

		if defaultTodoListID != "" {
			_, err := fmt.Sscanf(defaultTodoListID, "%d", &todoListID)
			if err != nil {
				return fmt.Errorf("invalid default todo list ID in config")
			}
		} else {
			return fmt.Errorf("no todo list specified. Use --list flag or run 'bc4 todo set' to set a default")
		}
	}

	// Create the todo group
	req := api.TodoGroupCreateRequest{
		Name: name,
	}

	group, err := todoOps.CreateTodoGroup(f.Context(), projectID, todoListID, req)
	if err != nil {
		return fmt.Errorf("failed to create todo group: %w", err)
	}

	// GitHub CLI style: minimal output with created ID
	fmt.Printf("Created todo group #%d\n", group.ID)

	return nil
}
