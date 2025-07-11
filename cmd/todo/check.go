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

func newCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check <todo-id>",
		Short: "Mark a todo as complete",
		Long: `Mark a todo as complete.

Provide the todo ID (with or without the # prefix) to mark it as done.`,
		Example: `  # Mark todo #12345 as complete
  bc4 todo check 12345

  # Also works with # prefix
  bc4 todo check #12345`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheck(cmd.Context(), args[0])
		},
	}

	return cmd
}

func runCheck(ctx context.Context, todoIDStr string) error {
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

	// Check if already completed
	if todo.Completed {
		fmt.Printf("✓ Todo #%d is already completed\n", todoID)
		return nil
	}

	// Mark as complete
	err = client.CompleteTodo(ctx, projectID, todoID)
	if err != nil {
		return fmt.Errorf("failed to complete todo: %w", err)
	}

	// GitHub CLI style: minimal output with confirmation
	fmt.Printf("✓ Completed #%d: %s\n", todoID, todo.Title)

	return nil
}
