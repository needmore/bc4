package todo

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
	"github.com/needmore/bc4/internal/ui"
)

func newViewCmd() *cobra.Command {
	var accountID string
	var projectID string
	var formatStr string
	var jsonFields string
	var webView bool

	cmd := &cobra.Command{
		Use:   "view [list-id|name]",
		Short: "View todos in a specific list",
		Long:  `View all todos in a specific todo list. Can specify by ID or partial name match.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Check if we have auth
			if cfg.ClientID == "" || cfg.ClientSecret == "" {
				return fmt.Errorf("not authenticated. Run 'bc4' to set up authentication")
			}

			// Create auth client
			authClient := auth.NewClient(cfg.ClientID, cfg.ClientSecret)

			// Use specified account or default
			if accountID == "" {
				accountID = authClient.GetDefaultAccount()
			}

			if accountID == "" {
				return fmt.Errorf("no account specified and no default account set")
			}

			// Use specified project or default
			if projectID == "" {
				projectID = cfg.DefaultProject
				if projectID == "" && cfg.Accounts != nil {
					if acc, ok := cfg.Accounts[accountID]; ok {
						projectID = acc.DefaultProject
					}
				}
			}

			if projectID == "" {
				return fmt.Errorf("no project specified and no default project set")
			}

			// Get token
			token, err := authClient.GetToken(accountID)
			if err != nil {
				return fmt.Errorf("failed to get auth token: %w", err)
			}

			// Create API client
			apiClient := api.NewClient(accountID, token.AccessToken)

			// Get todo set for the project
			todoSet, err := apiClient.GetProjectTodoSet(context.Background(), projectID)
			if err != nil {
				return fmt.Errorf("failed to get project todo set: %w", err)
			}

			// Determine which todo list to view
			var todoListID int64
			if len(args) == 0 {
				// No argument - use default todo list if set
				defaultTodoListID := ""
				if cfg.Accounts != nil && cfg.Accounts[accountID].ProjectDefaults != nil {
					if projDefaults, ok := cfg.Accounts[accountID].ProjectDefaults[projectID]; ok {
						defaultTodoListID = projDefaults.DefaultTodoList
					}
				}
				if defaultTodoListID == "" {
					return fmt.Errorf("no todo list specified and no default set. Use 'todo select' to set a default")
				}
				todoListID, _ = strconv.ParseInt(defaultTodoListID, 10, 64)
			} else {
				// Try to parse as ID first
				if id, err := strconv.ParseInt(args[0], 10, 64); err == nil {
					todoListID = id
				} else {
					// Try to find by name
					todoLists, err := apiClient.GetTodoLists(context.Background(), projectID, todoSet.ID)
					if err != nil {
						return fmt.Errorf("failed to fetch todo lists: %w", err)
					}

					searchTerm := strings.ToLower(args[0])
					var matches []api.TodoList
					for _, list := range todoLists {
						if strings.Contains(strings.ToLower(list.Title), searchTerm) {
							matches = append(matches, list)
						}
					}

					if len(matches) == 0 {
						return fmt.Errorf("no todo list found matching '%s'", args[0])
					} else if len(matches) > 1 {
						return fmt.Errorf("multiple todo lists match '%s'. Please be more specific", args[0])
					}

					todoListID = matches[0].ID
				}
			}

			// Get the todo list
			todoList, err := apiClient.GetTodoList(context.Background(), projectID, todoListID)
			if err != nil {
				return fmt.Errorf("failed to fetch todo list: %w", err)
			}

			// Handle web view
			if webView {
				// Open in browser
				url := fmt.Sprintf("https://3.basecamp.com/%s/buckets/%s/todolists/%d", accountID, projectID, todoListID)
				fmt.Printf("Opening %s in your browser...\n", url)
				// Note: In a real implementation, we'd use a cross-platform way to open URLs
				return nil
			}

			// Get todos in the list
			todos, err := apiClient.GetTodos(context.Background(), projectID, todoListID)
			if err != nil {
				return fmt.Errorf("failed to fetch todos: %w", err)
			}

			// Parse output format
			format, err := ui.ParseOutputFormat(formatStr)
			if err != nil {
				return err
			}

			// Handle JSON output
			if format == ui.OutputFormatJSON || jsonFields != "" {
				return outputTodoListJSON(todoList, todos, jsonFields)
			}

			// Display todo list in terminal
			return displayTodoList(todoList, todos, format)
		},
	}

	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID (overrides default)")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID (overrides default)")
	cmd.Flags().StringVarP(&formatStr, "format", "f", "table", "Output format: table, json, or tsv")
	cmd.Flags().StringVar(&jsonFields, "json", "", "Output JSON with specified fields")
	cmd.Flags().BoolVarP(&webView, "web", "w", false, "Open in web browser")

	return cmd
}

func displayTodoList(todoList *api.TodoList, todos []api.Todo, format ui.OutputFormat) error {
	// Terminal display with nice formatting
	if !ui.IsTerminal(os.Stdout) || format == ui.OutputFormatTSV {
		// Non-TTY or TSV format - simple output
		return displayTodoListSimple(todoList, todos)
	}

	// Pretty terminal display
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")).
		MarginBottom(1)

	metaStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	// Display title
	fmt.Println(titleStyle.Render(todoList.Title))

	// Display metadata
	completed := 0
	for _, todo := range todos {
		if todo.Completed {
			completed++
		}
	}

	meta := fmt.Sprintf("%d/%d completed", completed, len(todos))
	if todoList.CreatedAt != "" {
		if createdTime, err := time.Parse(time.RFC3339, todoList.CreatedAt); err == nil {
			meta += fmt.Sprintf(" • Created %s", createdTime.Format("Jan 2, 2006"))
		}
	}
	fmt.Println(metaStyle.Render(meta))
	fmt.Println()

	// Display description if present
	if todoList.Description != "" {
		// Render markdown description
		renderer, _ := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(ui.GetTerminalWidth()-4),
		)

		if rendered, err := renderer.Render(todoList.Description); err == nil {
			fmt.Print(rendered)
		} else {
			fmt.Println(todoList.Description)
		}
		fmt.Println()
	}

	// Display todos
	if len(todos) == 0 {
		fmt.Println(metaStyle.Render("No todos in this list"))
		return nil
	}

	// Create table for todos
	config := ui.NewOutputConfig(os.Stdout)
	config.Format = ui.OutputFormatTable
	tw := ui.NewTableWriter(config)

	// Add headers
	tw.AddHeader([]string{"", "Todo", "Due", "Assignee"})

	// Style for completed todos
	completedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	// Add todos
	for _, todo := range todos {
		status := "[ ]"
		if todo.Completed {
			status = "[✓]"
		}

		title := todo.Title
		if todo.Completed {
			title = completedStyle.Render(title)
		}

		due := ""
		if todo.DueOn != nil && *todo.DueOn != "" {
			if dueTime, err := time.Parse("2006-01-02", *todo.DueOn); err == nil {
				due = dueTime.Format("Jan 2")
			}
		}

		assignee := ""
		// Note: Assignee would need to be added to the Todo struct if available in API

		tw.AddRow([]string{status, title, due, assignee})
	}

	return tw.Render()
}

func displayTodoListSimple(todoList *api.TodoList, todos []api.Todo) error {
	// Simple TSV output
	fmt.Printf("Todo List: %s\n", todoList.Title)
	fmt.Printf("ID: %d\n", todoList.ID)

	completed := 0
	for _, todo := range todos {
		if todo.Completed {
			completed++
		}
	}
	fmt.Printf("Progress: %d/%d completed\n\n", completed, len(todos))

	// Output todos as TSV
	fmt.Println("Status\tTodo\tDue")
	for _, todo := range todos {
		status := "[ ]"
		if todo.Completed {
			status = "[x]"
		}

		due := ""
		if todo.DueOn != nil {
			due = *todo.DueOn
		}

		fmt.Printf("%s\t%s\t%s\n", status, todo.Title, due)
	}

	return nil
}

func outputTodoListJSON(todoList *api.TodoList, todos []api.Todo, fields string) error {
	// Combine todo list and todos data
	data := map[string]interface{}{
		"id":          todoList.ID,
		"title":       todoList.Title,
		"description": todoList.Description,
		"created_at":  todoList.CreatedAt,
		"updated_at":  todoList.UpdatedAt,
		"completed":   fmt.Sprintf("%d/%d", countCompleted(todos), len(todos)),
		"todos":       todos,
	}

	// If specific fields requested, filter the output
	if fields != "" {
		// This is a simplified version - in production you'd parse the fields
		// and extract only requested fields
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func countCompleted(todos []api.Todo) int {
	count := 0
	for _, todo := range todos {
		if todo.Completed {
			count++
		}
	}
	return count
}
