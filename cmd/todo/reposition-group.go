package todo

import (
	"fmt"
	"strconv"

	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
)

func newRepositionGroupCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reposition-group <group-id> <position>",
		Short: "Reposition a group within a todo list",
		Long: `Reposition a group (section) within a todo list to change its order.

The position is 1-based, where 1 is the first position in the list.
Groups are ordered from top to bottom in the todo list.`,
		Example: `  # Move a group to the first position
  bc4 todo reposition-group 12345 1

  # Move a group to the third position
  bc4 todo reposition-group 12345 3

  # Reposition a group using a Basecamp URL
  bc4 todo reposition-group https://3.basecamp.com/.../groups/12345 2`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRepositionGroup(f, args)
		},
	}

	return cmd
}

func runRepositionGroup(f *factory.Factory, args []string) error {
	// Parse group ID
	groupIDStr := args[0]
	var groupID int64
	var err error

	// Try to parse as URL first
	if parser.IsBasecampURL(groupIDStr) {
		parsed, err := parser.ParseBasecampURL(groupIDStr)
		if err != nil {
			return fmt.Errorf("invalid Basecamp URL: %w", err)
		}
		groupID = parsed.ResourceID
	} else {
		// Try to parse as numeric ID
		groupID, err = strconv.ParseInt(groupIDStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid group ID: %s (must be a numeric ID or Basecamp URL)", groupIDStr)
		}
	}

	// Parse position
	position, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid position: %s (must be a positive integer)", args[1])
	}

	if position < 1 {
		return fmt.Errorf("position must be at least 1")
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

	// Reposition the group
	err = todoOps.RepositionTodoGroup(f.Context(), projectID, groupID, position)
	if err != nil {
		return fmt.Errorf("failed to reposition todo group: %w", err)
	}

	// GitHub CLI style: minimal output
	fmt.Printf("Repositioned group #%d to position %d\n", groupID, position)

	return nil
}
