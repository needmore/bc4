package todo

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
	"github.com/spf13/cobra"
)

type addOptions struct {
	list        string
	description string
	due         string
	assign      []string
}

func newAddCmd() *cobra.Command {
	opts := &addOptions{}

	cmd := &cobra.Command{
		Use:   "add [<title>]",
		Short: "Create a new todo",
		Long: `Create a new todo in the specified todo list.

If no title is provided, you'll be prompted to enter one interactively.
The todo will be created in the default todo list unless specified with --list.`,
		Example: `  # Add a todo with a title
  bc4 todo add "Review pull request"

  # Add a todo with description
  bc4 todo add "Deploy to production" --description "After all tests pass"

  # Add a todo with due date
  bc4 todo add "Submit report" --due 2025-01-15

  # Add a todo to a specific list
  bc4 todo add "Update documentation" --list "Documentation Tasks"`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAdd(cmd.Context(), opts, args)
		},
	}

	cmd.Flags().StringVarP(&opts.list, "list", "l", "", "Todo list ID or name (defaults to selected list)")
	cmd.Flags().StringVarP(&opts.description, "description", "d", "", "Description for the todo")
	cmd.Flags().StringVar(&opts.due, "due", "", "Due date (YYYY-MM-DD)")
	cmd.Flags().StringSliceVar(&opts.assign, "assign", nil, "Assign to team members (by email)")

	return cmd
}

func runAdd(ctx context.Context, opts *addOptions, args []string) error {
	// Get title from args or prompt
	var title string
	if len(args) > 0 {
		title = args[0]
	} else {
		// TODO: Add interactive prompt using Bubbletea
		return fmt.Errorf("interactive mode not yet implemented. Please provide a title as an argument")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get authentication token
	authClient := auth.NewClient(cfg.ClientID, cfg.ClientSecret)
	token, err := authClient.GetToken(cfg.DefaultAccount)
	if err != nil {
		return fmt.Errorf("not authenticated. Run 'bc4 auth login' first")
	}

	// Get current project
	projectID := cfg.DefaultProject
	if projectID == "" && cfg.Accounts != nil {
		if acc, ok := cfg.Accounts[cfg.DefaultAccount]; ok {
			projectID = acc.DefaultProject
		}
	}
	if projectID == "" {
		return fmt.Errorf("no project selected. Run 'bc4 project select' first")
	}

	// Create API client
	client := api.NewClient(cfg.DefaultAccount, token.AccessToken)

	// Get the todo set for the project
	todoSet, err := client.GetProjectTodoSet(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get todo set: %w", err)
	}

	// Determine which todo list to use
	var todoListID int64
	if opts.list != "" {
		// User specified a list - try to find it
		todoLists, err := client.GetTodoLists(ctx, projectID, todoSet.ID)
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
	} else {
		// Use default todo list from config
		defaultTodoListID := ""
		if cfg.Accounts != nil {
			if acc, ok := cfg.Accounts[cfg.DefaultAccount]; ok {
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

	// Create the todo
	req := api.TodoCreateRequest{
		Content:     title,
		Description: opts.description,
	}

	if opts.due != "" {
		req.DueOn = &opts.due
	}

	// TODO: Handle assignee lookup by email
	if len(opts.assign) > 0 {
		fmt.Fprintln(os.Stderr, "Warning: --assign flag not yet implemented")
	}

	todo, err := client.CreateTodo(ctx, projectID, todoListID, req)
	if err != nil {
		return fmt.Errorf("failed to create todo: %w", err)
	}

	// Output the created todo ID (GitHub CLI style - minimal output)
	fmt.Printf("#%d\n", todo.ID)

	return nil
}
