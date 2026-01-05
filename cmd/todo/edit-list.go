package todo

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/cmdutil"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/markdown"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
)

type editListOptions struct {
	name             string
	description      string
	clearDescription bool
}

func newEditListCmd(f *factory.Factory) *cobra.Command {
	opts := &editListOptions{}

	cmd := &cobra.Command{
		Use:   "edit-list <list-id|name|url>",
		Short: "Edit an existing todo list",
		Long: `Edit an existing todo list's name or description.

You can specify the todo list using:
- A numeric ID (e.g., "12345")
- A partial name match (e.g., "Sprint Tasks")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/todolists/12345")

At least one of --name or --description must be specified.
The name field is required by the API, so if only updating description, the current name is preserved.`,
		Example: `  # Rename a todo list
  bc4 todo edit-list 12345 --name "Renamed List"

  # Update description with markdown
  bc4 todo edit-list 12345 --description "Updated description with **bold** text"

  # Rename and update description
  bc4 todo edit-list "Sprint Tasks" --name "Sprint 1" --description "First sprint tasks"

  # Clear description
  bc4 todo edit-list 12345 --clear-description

  # Edit using a URL
  bc4 todo edit-list https://3.basecamp.com/.../todolists/12345 --name "New Name"`,
		Args: cmdutil.ExactArgs(1, "list-id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEditList(f, opts, args)
		},
	}

	cmd.Flags().StringVarP(&opts.name, "name", "n", "", "New name for the todo list")
	cmd.Flags().StringVarP(&opts.description, "description", "d", "", "New description for the todo list (supports markdown)")
	cmd.Flags().BoolVar(&opts.clearDescription, "clear-description", false, "Clear the description")

	return cmd
}

func runEditList(f *factory.Factory, opts *editListOptions, args []string) error {
	// Check if any changes were requested
	if opts.name == "" && opts.description == "" && !opts.clearDescription {
		return fmt.Errorf("no changes specified. Use --name, --description, or --clear-description")
	}

	// Parse list ID (could be numeric ID, name, or URL)
	listArg := args[0]
	var todoListID int64
	var err error

	// Try to parse as URL first
	if parser.IsBasecampURL(listArg) {
		parsed, err := parser.ParseBasecampURL(listArg)
		if err != nil {
			return fmt.Errorf("invalid Basecamp URL: %w", err)
		}
		todoListID = parsed.ResourceID
		if parsed.AccountID > 0 {
			f = f.WithAccount(strconv.FormatInt(parsed.AccountID, 10))
		}
		if parsed.ProjectID > 0 {
			f = f.WithProject(strconv.FormatInt(parsed.ProjectID, 10))
		}
	} else if id, parseErr := strconv.ParseInt(listArg, 10, 64); parseErr == nil {
		// Numeric ID
		todoListID = id
	} else {
		// Try to find by name - we'll do this after getting the API client
		todoListID = 0
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

	// If we need to find by name, do it now
	if todoListID == 0 {
		todoSet, err := todoOps.GetProjectTodoSet(f.Context(), projectID)
		if err != nil {
			return fmt.Errorf("failed to get todo set: %w", err)
		}

		todoLists, err := todoOps.GetTodoLists(f.Context(), projectID, todoSet.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch todo lists: %w", err)
		}

		searchTerm := strings.ToLower(listArg)
		var matches []api.TodoList
		for _, list := range todoLists {
			if strings.Contains(strings.ToLower(list.Title), searchTerm) {
				matches = append(matches, list)
			}
		}

		if len(matches) == 0 {
			return fmt.Errorf("no todo list found matching '%s'", listArg)
		} else if len(matches) > 1 {
			return fmt.Errorf("multiple todo lists match '%s'. Please be more specific or use the list ID", listArg)
		}

		todoListID = matches[0].ID
	}

	// Fetch the current todo list to get existing values
	currentList, err := todoOps.GetTodoList(f.Context(), projectID, todoListID)
	if err != nil {
		return fmt.Errorf("failed to get todo list: %w", err)
	}

	// Build update request - name is required by the API
	req := api.TodoListUpdateRequest{
		Name: currentList.Title, // Default to current name
	}

	// Update name if specified
	if opts.name != "" {
		req.Name = opts.name
	}

	// Handle description update
	if opts.clearDescription {
		req.Description = ""
	} else if opts.description != "" {
		// Convert markdown to rich text
		converter := markdown.NewConverter()
		richDescription, err := converter.MarkdownToRichText(opts.description)
		if err != nil {
			return fmt.Errorf("failed to convert description: %w", err)
		}
		req.Description = richDescription
	} else {
		// Keep current description
		req.Description = currentList.Description
	}

	// Update the todo list
	updatedList, err := todoOps.UpdateTodoList(f.Context(), projectID, todoListID, req)
	if err != nil {
		return fmt.Errorf("failed to update todo list: %w", err)
	}

	// GitHub CLI style: minimal output
	fmt.Printf("Updated todo list #%d\n", updatedList.ID)

	return nil
}
