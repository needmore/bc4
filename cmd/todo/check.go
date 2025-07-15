package todo

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
)

func newCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check <todo-id or URL>",
		Short: "Mark a todo as complete",
		Long: `Mark a todo as complete.

You can specify the todo using either:
- A numeric ID (e.g., "12345" or "#12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/todos/12345")`,
		Example: `  # Mark todo #12345 as complete
  bc4 todo check 12345

  # Also works with # prefix
  bc4 todo check #12345

  # Using a Basecamp URL
  bc4 todo check "https://3.basecamp.com/1234567/buckets/89012345/todos/12345"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheck(cmd.Context(), args[0])
		},
	}

	return cmd
}

func runCheck(ctx context.Context, todoIDStr string) error {
	// Load configuration first
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Parse todo ID (handle #123 format and URLs)
	todoIDStr = strings.TrimPrefix(todoIDStr, "#")
	todoID, parsedURL, err := parser.ParseArgument(todoIDStr)
	if err != nil {
		return fmt.Errorf("invalid todo ID or URL: %s", todoIDStr)
	}

	// Initialize account and project IDs from config
	accountID := cfg.DefaultAccount
	projectID := cfg.DefaultProject
	if cfg.Accounts != nil {
		if acc, ok := cfg.Accounts[accountID]; ok && acc.DefaultProject != "" {
			projectID = acc.DefaultProject
		}
	}

	// If a URL was parsed, override account and project IDs if provided
	if parsedURL != nil {
		if parsedURL.ResourceType != parser.ResourceTypeTodo {
			return fmt.Errorf("URL is not for a todo: %s", todoIDStr)
		}
		if parsedURL.AccountID > 0 {
			accountID = strconv.FormatInt(parsedURL.AccountID, 10)
		}
		if parsedURL.ProjectID > 0 {
			projectID = strconv.FormatInt(parsedURL.ProjectID, 10)
		}
	}

	// Validate we have required IDs
	if accountID == "" {
		return fmt.Errorf("no account specified. Run 'bc4 account select' first")
	}
	if projectID == "" {
		return fmt.Errorf("no project specified. Run 'bc4 project select' first")
	}

	// Get authentication token
	authClient := auth.NewClient(cfg.ClientID, cfg.ClientSecret)
	token, err := authClient.GetToken(accountID)
	if err != nil {
		return fmt.Errorf("not authenticated. Run 'bc4 auth login' first")
	}

	// Create API client
	client := api.NewClient(accountID, token.AccessToken)

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
