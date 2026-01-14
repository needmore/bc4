package todo

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/attachments"
	"github.com/needmore/bc4/internal/cmdutil"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/markdown"
	"github.com/needmore/bc4/internal/parser"
	"github.com/needmore/bc4/internal/utils"
	"github.com/spf13/cobra"
)

type editOptions struct {
	title       string
	description string
	due         string
	startsOn    string
	assign      []string
	unassign    []string
	file        string
	clearDue    bool
	attach      []string
}

func newEditCmd(f *factory.Factory) *cobra.Command {
	opts := &editOptions{}

	cmd := &cobra.Command{
		Use:   "edit <todo-id|url>",
		Short: "Edit an existing todo",
		Long: `Edit an existing todo's title, description, due date, or assignees.

You can specify the todo using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/todos/12345")

All fields are optional - only specified fields will be updated.

Use --attach to add images or files to the todo description. Attachments are
appended to the existing description. Multiple files can be attached by using
the flag multiple times.`,
		Example: `  # Edit todo title
  bc4 todo edit 12345 --title "Updated title"

  # Edit todo description with markdown
  bc4 todo edit 12345 --description "New description with **bold** text"

  # Set due date
  bc4 todo edit 12345 --due 2025-02-15

  # Clear due date
  bc4 todo edit 12345 --clear-due

  # Assign someone to the todo
  bc4 todo edit 12345 --assign user@example.com

  # Remove someone from the todo
  bc4 todo edit 12345 --unassign user@example.com

  # Update from a markdown file
  bc4 todo edit 12345 --file updated-todo.md

  # Edit using a URL
  bc4 todo edit https://3.basecamp.com/.../todos/12345 --title "New title"

  # Add an image attachment to an existing todo
  bc4 todo edit 12345 --attach ./screenshot.png

  # Add multiple attachments
  bc4 todo edit 12345 --attach ./photo1.jpg --attach ./photo2.jpg`,
		Args: cmdutil.ExactArgs(1, "todo-id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEdit(f, opts, args)
		},
	}

	cmd.Flags().StringVarP(&opts.title, "title", "t", "", "New title for the todo")
	cmd.Flags().StringVarP(&opts.description, "description", "d", "", "New description for the todo")
	cmd.Flags().StringVar(&opts.due, "due", "", "Due date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&opts.startsOn, "starts-on", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().StringSliceVar(&opts.assign, "assign", nil, "Add assignees (by email or name)")
	cmd.Flags().StringSliceVar(&opts.unassign, "unassign", nil, "Remove assignees (by email or name)")
	cmd.Flags().StringVarP(&opts.file, "file", "f", "", "Read new content from a markdown file")
	cmd.Flags().BoolVar(&opts.clearDue, "clear-due", false, "Clear the due date")
	cmd.Flags().StringSliceVar(&opts.attach, "attach", nil, "Attach file(s) to the todo (can be used multiple times)")

	return cmd
}

func runEdit(f *factory.Factory, opts *editOptions, args []string) error {
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

	// Get API client from factory
	client, err := f.ApiClient()
	if err != nil {
		return err
	}
	todoOps := client.Todos()

	// Get resolved project ID
	resolvedProjectID, err := f.ProjectID()
	if err != nil {
		return err
	}

	// Fetch the current todo to get existing values
	currentTodo, err := todoOps.GetTodo(f.Context(), resolvedProjectID, todoID)
	if err != nil {
		return fmt.Errorf("failed to get todo: %w", err)
	}

	// Handle file input
	var fileContent string
	if opts.file != "" {
		data, err := os.ReadFile(opts.file)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		fileContent = strings.TrimSpace(string(data))
	} else {
		// Check if stdin has data
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			// Data is being piped in
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read from stdin: %w", err)
			}
			fileContent = strings.TrimSpace(string(data))
		}
	}

	// If file content provided, parse title and description from it
	if fileContent != "" {
		lines := strings.SplitN(fileContent, "\n", 2)
		if opts.title == "" {
			opts.title = strings.TrimSpace(lines[0])
		}
		if len(lines) > 1 && opts.description == "" {
			opts.description = strings.TrimSpace(lines[1])
		}
	}

	// Check if any changes were requested
	hasChanges := opts.title != "" || opts.description != "" || opts.due != "" ||
		opts.startsOn != "" || len(opts.assign) > 0 || len(opts.unassign) > 0 ||
		opts.clearDue || len(opts.attach) > 0

	if !hasChanges {
		return fmt.Errorf("no changes specified. Use --title, --description, --due, --assign, --unassign, --attach, or --file to specify changes")
	}

	// Build update request
	req := api.TodoUpdateRequest{}

	// Create markdown converter
	converter := markdown.NewConverter()

	// Handle title update
	if opts.title != "" {
		richTitle, err := converter.MarkdownToRichText(opts.title)
		if err != nil {
			return fmt.Errorf("failed to convert title: %w", err)
		}
		req.Content = richTitle
	}

	// Handle description update
	if opts.description != "" {
		richDescription, err := converter.MarkdownToRichText(opts.description)
		if err != nil {
			return fmt.Errorf("failed to convert description: %w", err)
		}
		req.Description = richDescription
	}

	// Handle attachments - append to existing or new description
	if len(opts.attach) > 0 {
		// Start with the description we already have, or the current todo's description
		baseDescription := req.Description
		if baseDescription == "" {
			baseDescription = currentTodo.Description
		}

		for _, attachPath := range opts.attach {
			fileData, err := os.ReadFile(attachPath)
			if err != nil {
				return fmt.Errorf("failed to read attachment %s: %w", attachPath, err)
			}
			filename := filepath.Base(attachPath)
			upload, err := client.UploadAttachment(filename, fileData, "")
			if err != nil {
				return fmt.Errorf("failed to upload attachment %s: %w", filename, err)
			}
			tag := attachments.BuildTag(upload.AttachableSGID)
			baseDescription += tag
		}
		req.Description = baseDescription
	}

	// Handle due date
	if opts.clearDue {
		emptyDate := ""
		req.DueOn = &emptyDate
	} else if opts.due != "" {
		req.DueOn = &opts.due
	}

	// Handle start date
	if opts.startsOn != "" {
		req.StartsOn = &opts.startsOn
	}

	// Handle assignee changes
	if len(opts.assign) > 0 || len(opts.unassign) > 0 {
		// Create user resolver
		userResolver := utils.NewUserResolver(client.Client, resolvedProjectID)

		// Start with existing assignees
		currentAssigneeIDs := make([]int64, 0)
		for _, assignee := range currentTodo.Assignees {
			currentAssigneeIDs = append(currentAssigneeIDs, assignee.ID)
		}

		// Add new assignees
		if len(opts.assign) > 0 {
			newAssigneeIDs, err := userResolver.ResolveUsers(f.Context(), opts.assign)
			if err != nil {
				return fmt.Errorf("failed to resolve assignees to add: %w", err)
			}
			// Merge without duplicates
			for _, newID := range newAssigneeIDs {
				found := false
				for _, existingID := range currentAssigneeIDs {
					if existingID == newID {
						found = true
						break
					}
				}
				if !found {
					currentAssigneeIDs = append(currentAssigneeIDs, newID)
				}
			}
		}

		// Remove assignees
		if len(opts.unassign) > 0 {
			removeIDs, err := userResolver.ResolveUsers(f.Context(), opts.unassign)
			if err != nil {
				return fmt.Errorf("failed to resolve assignees to remove: %w", err)
			}
			// Filter out removed assignees
			filteredIDs := make([]int64, 0)
			for _, existingID := range currentAssigneeIDs {
				shouldRemove := false
				for _, removeID := range removeIDs {
					if existingID == removeID {
						shouldRemove = true
						break
					}
				}
				if !shouldRemove {
					filteredIDs = append(filteredIDs, existingID)
				}
			}
			currentAssigneeIDs = filteredIDs
		}

		req.AssigneeIDs = currentAssigneeIDs
	}

	// Update the todo
	updatedTodo, err := todoOps.UpdateTodo(f.Context(), resolvedProjectID, todoID, req)
	if err != nil {
		return fmt.Errorf("failed to update todo: %w", err)
	}

	// Output the updated todo ID (GitHub CLI style - minimal output)
	fmt.Printf("Updated #%d\n", updatedTodo.ID)

	return nil
}
