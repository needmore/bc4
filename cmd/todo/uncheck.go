package todo

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
	"github.com/spf13/cobra"
)

func newUncheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uncheck <todo-id>",
		Short: "Mark a todo as incomplete",
		Long: `Mark a todo as incomplete.

Provide the todo ID (with or without the # prefix) to mark it as not done.`,
		Example: `  # Mark todo #12345 as incomplete
  bc4 todo uncheck 12345

  # Also works with # prefix
  bc4 todo uncheck #12345`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUncheck(cmd.Context(), args[0])
		},
	}

	return cmd
}

func runUncheck(ctx context.Context, todoIDStr string) error {
	// Parse todo ID (handle #123 format)
	todoIDStr = strings.TrimPrefix(todoIDStr, "#")
	todoID, err := strconv.ParseInt(todoIDStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid todo ID: %s", todoIDStr)
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

	// Get the todo first to display its title
	todo, err := client.GetTodo(ctx, projectID, todoID)
	if err != nil {
		return fmt.Errorf("failed to fetch todo: %w", err)
	}

	// Check if already incomplete
	if !todo.Completed {
		fmt.Printf("○ Todo #%d is already incomplete\n", todoID)
		return nil
	}

	// Mark as incomplete
	err = client.UncompleteTodo(ctx, projectID, todoID)
	if err != nil {
		return fmt.Errorf("failed to uncomplete todo: %w", err)
	}

	// GitHub CLI style: minimal output with confirmation
	fmt.Printf("○ Reopened #%d: %s\n", todoID, todo.Title)

	return nil
}
