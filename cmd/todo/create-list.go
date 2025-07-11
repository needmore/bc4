package todo

import (
	"context"
	"fmt"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
	"github.com/spf13/cobra"
)

type createListOptions struct {
	description string
}

func newCreateListCmd() *cobra.Command {
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
			return runCreateList(cmd.Context(), opts, args)
		},
	}

	cmd.Flags().StringVarP(&opts.description, "description", "d", "", "Description for the todo list")

	return cmd
}

func runCreateList(ctx context.Context, opts *createListOptions, args []string) error {
	// Get name from args or prompt
	var name string
	if len(args) > 0 {
		name = args[0]
	} else {
		// TODO: Add interactive prompt using Bubbletea
		return fmt.Errorf("interactive mode not yet implemented. Please provide a name as an argument")
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

	// Create the todo list
	req := api.TodoListCreateRequest{
		Name:        name,
		Description: opts.description,
	}

	todoList, err := client.CreateTodoList(ctx, projectID, todoSet.ID, req)
	if err != nil {
		return fmt.Errorf("failed to create todo list: %w", err)
	}

	// GitHub CLI style: minimal output with created ID
	fmt.Printf("Created todo list #%d\n", todoList.ID)

	// Suggest setting as default
	fmt.Printf("\nTo set as default: bc4 todo set %d\n", todoList.ID)

	return nil
}
