package todo

import (
	"fmt"
	"strconv"

	"github.com/needmore/bc4/internal/cmdutil"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
)

type moveOptions struct {
	position int
	top      bool
	bottom   bool
}

func newMoveCmd(f *factory.Factory) *cobra.Command {
	opts := &moveOptions{}

	cmd := &cobra.Command{
		Use:   "move <todo-id|url>",
		Short: "Move a todo to a different position within its list",
		Long: `Move a todo to a different position within its todo list.

You can specify the todo using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/todos/12345")

Position is 1-based (1 = first item in the list).`,
		Example: `  # Move todo to specific position
  bc4 todo move 12345 --position 1      # Move to top (first position)
  bc4 todo move 12345 --position 3      # Move to 3rd position

  # Move to top or bottom
  bc4 todo move 12345 --top             # Move to top of list
  bc4 todo move 12345 --bottom          # Move to bottom of list

  # Move using a URL
  bc4 todo move https://3.basecamp.com/.../todos/12345 --position 1`,
		Args: cmdutil.ExactArgs(1, "todo-id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMove(f, opts, args)
		},
	}

	cmd.Flags().IntVar(&opts.position, "position", 0, "Move to specific position (1-based)")
	cmd.Flags().BoolVar(&opts.top, "top", false, "Move to top of list (position 1)")
	cmd.Flags().BoolVar(&opts.bottom, "bottom", false, "Move to bottom of list")

	return cmd
}

func runMove(f *factory.Factory, opts *moveOptions, args []string) error {
	// Parse todo ID (could be numeric ID or URL)
	todoID, parsedURL, err := parser.ParseArgument(args[0])
	if err != nil {
		return fmt.Errorf("invalid todo ID or URL: %s", args[0])
	}

	// If a URL was parsed, override account and project IDs if provided
	if parsedURL != nil {
		if parsedURL.ResourceType != parser.ResourceTypeTodo {
			return fmt.Errorf("URL is not for a todo: %s", args[0])
		}
		if parsedURL.AccountID > 0 {
			f = f.WithAccount(strconv.FormatInt(parsedURL.AccountID, 10))
		}
		if parsedURL.ProjectID > 0 {
			f = f.WithProject(strconv.FormatInt(parsedURL.ProjectID, 10))
		}
	}

	// Validate position options
	optionCount := 0
	if opts.position > 0 {
		optionCount++
	}
	if opts.top {
		optionCount++
	}
	if opts.bottom {
		optionCount++
	}

	if optionCount == 0 {
		return fmt.Errorf("specify a position using --position, --top, or --bottom")
	}
	if optionCount > 1 {
		return fmt.Errorf("only one of --position, --top, or --bottom can be specified")
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

	// Determine position
	var position int
	var positionLabel string

	if opts.top {
		position = 1
		positionLabel = "top"
	} else if opts.bottom {
		// Get the todo to find its list, then count todos to find bottom position
		todo, err := todoOps.GetTodo(f.Context(), projectID, todoID)
		if err != nil {
			return fmt.Errorf("failed to get todo: %w", err)
		}

		todos, err := todoOps.GetAllTodos(f.Context(), projectID, todo.TodolistID)
		if err != nil {
			return fmt.Errorf("failed to get todos in list: %w", err)
		}

		position = len(todos)
		positionLabel = "bottom"
	} else {
		position = opts.position
		positionLabel = fmt.Sprintf("position %d", position)
	}

	if position < 1 {
		return fmt.Errorf("position must be at least 1")
	}

	// Reposition the todo
	err = todoOps.RepositionTodo(f.Context(), projectID, todoID, position)
	if err != nil {
		return fmt.Errorf("failed to move todo: %w", err)
	}

	// GitHub CLI style: minimal output
	fmt.Printf("Moved #%d to %s\n", todoID, positionLabel)

	return nil
}
