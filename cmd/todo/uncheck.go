package todo

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
)

func newUncheckCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uncheck <todo-id or URL>",
		Short: "Mark a todo as incomplete",
		Long: `Mark a todo as incomplete.

You can specify the todo using either:
- A numeric ID (e.g., "12345" or "#12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/todos/12345")`,
		Example: `  # Mark todo #12345 as incomplete
  bc4 todo uncheck 12345

  # Also works with # prefix
  bc4 todo uncheck #12345

  # Using a Basecamp URL
  bc4 todo uncheck "https://3.basecamp.com/1234567/buckets/89012345/todos/12345"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUncheck(f, args[0])
		},
	}

	return cmd
}

func runUncheck(f *factory.Factory, todoIDStr string) error {
	// Parse todo ID (handle #123 format and URLs)
	todoIDStr = strings.TrimPrefix(todoIDStr, "#")
	todoID, parsedURL, err := parser.ParseArgument(todoIDStr)
	if err != nil {
		return fmt.Errorf("invalid todo ID or URL: %s", todoIDStr)
	}

	// If a URL was parsed, override account and project IDs if provided
	if parsedURL != nil {
		if parsedURL.ResourceType != parser.ResourceTypeTodo {
			return fmt.Errorf("URL is not for a todo: %s", todoIDStr)
		}
		if parsedURL.AccountID > 0 {
			f = f.WithAccount(strconv.FormatInt(parsedURL.AccountID, 10))
		}
		if parsedURL.ProjectID > 0 {
			f = f.WithProject(strconv.FormatInt(parsedURL.ProjectID, 10))
		}
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

	// Get the todo first to display its title
	todo, err := todoOps.GetTodo(f.Context(), projectID, todoID)
	if err != nil {
		return fmt.Errorf("failed to fetch todo: %w", err)
	}

	// Check if already incomplete
	if !todo.Completed {
		fmt.Printf("○ Todo #%d is already incomplete\n", todoID)
		return nil
	}

	// Mark as incomplete
	err = todoOps.UncompleteTodo(f.Context(), projectID, todoID)
	if err != nil {
		return fmt.Errorf("failed to uncomplete todo: %w", err)
	}

	// GitHub CLI style: minimal output with confirmation
	fmt.Printf("○ Reopened #%d: %s\n", todoID, todo.Title)

	return nil
}
