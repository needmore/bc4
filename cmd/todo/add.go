package todo

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/attachments"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/markdown"
	"github.com/needmore/bc4/internal/parser"
	"github.com/needmore/bc4/internal/utils"
	"github.com/spf13/cobra"
)

type addOptions struct {
	list        string
	description string
	due         string
	assign      []string
	file        string
	attach      []string
}

func newAddCmd(f *factory.Factory) *cobra.Command {
	opts := &addOptions{}

	cmd := &cobra.Command{
		Use:   "add [<title>]",
		Short: "Create a new todo",
		Long: `Create a new todo in the specified todo list.

If no title is provided, you'll be prompted to enter one interactively.
The todo will be created in the default todo list unless specified with --list.

Use --attach to add images or files to the todo description. Multiple files
can be attached by using the flag multiple times.`,
		Example: `  # Add a todo with a title
  bc4 todo add "Review pull request"

  # Add a todo with description
  bc4 todo add "Deploy to production" --description "After all tests pass"

  # Add a todo with due date
  bc4 todo add "Submit report" --due 2025-01-15

  # Add a todo to a specific list
  bc4 todo add "Update documentation" --list "Documentation Tasks"

  # Add a todo from a markdown file
  bc4 todo add --file todo-content.md

  # Add a todo with an image attachment
  bc4 todo add "Review mockups" --attach ./design.png

  # Add a todo with multiple attachments
  bc4 todo add "Update assets" --attach ./image1.png --attach ./image2.jpg`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAdd(f, opts, args)
		},
	}

	cmd.Flags().StringVarP(&opts.list, "list", "l", "", "Todo list ID, name, or URL (defaults to selected list)")
	cmd.Flags().StringVarP(&opts.description, "description", "d", "", "Description for the todo")
	cmd.Flags().StringVar(&opts.due, "due", "", "Due date (YYYY-MM-DD)")
	cmd.Flags().StringSliceVar(&opts.assign, "assign", nil, "Assign to team members (by email)")
	cmd.Flags().StringVarP(&opts.file, "file", "f", "", "Read todo content from a markdown file")
	cmd.Flags().StringSliceVar(&opts.attach, "attach", nil, "Attach file(s) to the todo (can be used multiple times)")

	return cmd
}

func runAdd(f *factory.Factory, opts *addOptions, args []string) error {
	// Get content from file, stdin, args, or prompt
	var content string
	var err error

	if opts.file != "" {
		// Read from file
		data, err := os.ReadFile(opts.file)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		content = string(data)
	} else if len(args) > 0 {
		// Use argument as content
		content = args[0]
	} else {
		// Check if stdin has data
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			// Data is being piped in
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read from stdin: %w", err)
			}
			content = string(data)
		} else {
			// TODO: Add interactive prompt using Bubbletea
			return fmt.Errorf("interactive mode not yet implemented. Please provide content as an argument, via --file, or pipe it in")
		}
	}

	content = strings.TrimSpace(content)
	if content == "" {
		return fmt.Errorf("todo content cannot be empty")
	}

	// Split content into title and description if it's multi-line
	var title, description string
	lines := strings.SplitN(content, "\n", 2)
	title = strings.TrimSpace(lines[0])
	if len(lines) > 1 && opts.description == "" {
		description = strings.TrimSpace(lines[1])
	} else {
		description = opts.description
	}

	// Get API client from factory
	client, err := f.ApiClient()
	if err != nil {
		return err
	}
	todoOps := client.Todos()

	// Get config for default lookups
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	// Get resolved account ID
	resolvedAccountID, err := f.AccountID()
	if err != nil {
		return err
	}

	// Get resolved project ID
	resolvedProjectID, err := f.ProjectID()
	if err != nil {
		return err
	}

	// Get the todo set for the project
	todoSet, err := todoOps.GetProjectTodoSet(f.Context(), resolvedProjectID)
	if err != nil {
		return fmt.Errorf("failed to get todo set: %w", err)
	}

	// Determine which todo list to use
	var todoListID int64
	if opts.list != "" {
		// Check if it's a URL
		if parser.IsBasecampURL(opts.list) {
			parsed, err := parser.ParseBasecampURL(opts.list)
			if err != nil {
				return fmt.Errorf("invalid Basecamp URL: %w", err)
			}
			if parsed.ResourceType != parser.ResourceTypeTodoList {
				return fmt.Errorf("URL is not a todo list URL: %s", opts.list)
			}
			todoListID = parsed.ResourceID
		} else {
			// User specified a list - try to find it
			todoLists, err := todoOps.GetTodoLists(f.Context(), resolvedProjectID, todoSet.ID)
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
		}
	} else {
		// Use default todo list from config
		defaultTodoListID := ""
		if cfg.Accounts != nil {
			if acc, ok := cfg.Accounts[resolvedAccountID]; ok {
				if acc.ProjectDefaults != nil {
					if projDefaults, ok := acc.ProjectDefaults[resolvedProjectID]; ok {
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

	// Create markdown converter
	converter := markdown.NewConverter()

	// Convert title to rich text
	richTitle, err := converter.MarkdownToRichText(title)
	if err != nil {
		return fmt.Errorf("failed to convert title: %w", err)
	}

	// Convert description to rich text if provided
	var richDescription string
	if description != "" {
		richDescription, err = converter.MarkdownToRichText(description)
		if err != nil {
			return fmt.Errorf("failed to convert description: %w", err)
		}
	}

	// Handle attachments
	if len(opts.attach) > 0 {
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
			richDescription += tag
		}
	}

	// Create the todo
	req := api.TodoCreateRequest{
		Content:     richTitle,
		Description: richDescription,
	}

	if opts.due != "" {
		req.DueOn = &opts.due
	}

	// Handle assignee lookup
	if len(opts.assign) > 0 {
		// Create user resolver
		userResolver := utils.NewUserResolver(client.Client, resolvedProjectID)

		// Resolve user identifiers to person IDs
		personIDs, err := userResolver.ResolveUsers(f.Context(), opts.assign)
		if err != nil {
			return fmt.Errorf("failed to resolve assignees: %w", err)
		}

		req.AssigneeIDs = personIDs
	}

	todo, err := todoOps.CreateTodo(f.Context(), resolvedProjectID, todoListID, req)
	if err != nil {
		return fmt.Errorf("failed to create todo: %w", err)
	}

	// Output the created todo ID (GitHub CLI style - minimal output)
	fmt.Printf("#%d\n", todo.ID)

	return nil
}
