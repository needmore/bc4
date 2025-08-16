package todo

import (
	"encoding/csv"
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
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/ui"
	"github.com/needmore/bc4/internal/ui/tableprinter"
)

func newListCmd(f *factory.Factory) *cobra.Command {
	var accountID string
	var projectID string
	var formatStr string
	var jsonFields string
	var webView bool
	var showAll bool
	var grouped bool

	cmd := &cobra.Command{
		Use:   "list [list-id|name]",
		Short: "View todos in a specific list",
		Long: `View all todos in a specific todo list. Can specify by ID or partial name match.

For todo lists that are organized into groups/sections, use --grouped to display
them with clear section headers, or leave it off to show all todos in a flat table
with a GROUP column for easy scanning.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Apply account override if specified
			if accountID != "" {
				f = f.WithAccount(accountID)
			}

			// Apply project override if specified
			if projectID != "" {
				f = f.WithProject(projectID)
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

			// Get todo set for the project
			todoSet, err := todoOps.GetProjectTodoSet(f.Context(), resolvedProjectID)
			if err != nil {
				return fmt.Errorf("failed to get project todo set: %w", err)
			}

			// Determine which todo list to view
			var todoListID int64
			if len(args) == 0 {
				// No argument - use default todo list if set
				defaultTodoListID := ""
				if cfg.Accounts != nil && cfg.Accounts[resolvedAccountID].ProjectDefaults != nil {
					if projDefaults, ok := cfg.Accounts[resolvedAccountID].ProjectDefaults[resolvedProjectID]; ok {
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
					todoLists, err := todoOps.GetTodoLists(f.Context(), resolvedProjectID, todoSet.ID)
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
			todoList, err := todoOps.GetTodoList(f.Context(), resolvedProjectID, todoListID)
			if err != nil {
				return fmt.Errorf("failed to fetch todo list: %w", err)
			}

			// Handle web view
			if webView {
				// Open in browser
				url := fmt.Sprintf("https://3.basecamp.com/%s/buckets/%s/todolists/%d", resolvedAccountID, resolvedProjectID, todoListID)
				fmt.Printf("Opening %s in your browser...\n", url)
				// Note: In a real implementation, we'd use a cross-platform way to open URLs
				return nil
			}

			// Get todos in the list
			var todos []api.Todo
			if showAll {
				todos, err = todoOps.GetAllTodos(f.Context(), resolvedProjectID, todoListID)
			} else {
				todos, err = todoOps.GetTodos(f.Context(), resolvedProjectID, todoListID)
			}
			if err != nil {
				return fmt.Errorf("failed to fetch todos: %w", err)
			}

			// Check if this todo list has groups instead of direct todos
			var groups []api.TodoGroup
			var groupedTodos map[string][]api.Todo

			if len(todos) == 0 && todoList.GroupsURL != "" {
				// Try fetching groups
				groups, err = todoOps.GetTodoGroups(f.Context(), resolvedProjectID, todoListID)
				if err == nil && len(groups) > 0 {
					// Fetch todos for each group
					groupedTodos = make(map[string][]api.Todo)
					for _, group := range groups {
						var groupTodos []api.Todo
						if showAll {
							groupTodos, err = todoOps.GetAllTodos(f.Context(), resolvedProjectID, group.ID)
						} else {
							groupTodos, err = todoOps.GetTodos(f.Context(), resolvedProjectID, group.ID)
						}
						if err == nil {
							groupedTodos[fmt.Sprintf("%d", group.ID)] = groupTodos
						}
					}
				}
			}

			// Parse output format
			format, err := ui.ParseOutputFormat(formatStr)
			if err != nil {
				return err
			}

			// Handle JSON output
			if format == ui.OutputFormatJSON || jsonFields != "" {
				if len(groups) > 0 {
					return outputTodoListWithGroupsJSON(todoList, groups, groupedTodos, jsonFields)
				}
				return outputTodoListJSON(todoList, todos, jsonFields)
			}

			// Display todo list in terminal - GitHub CLI style
			if len(groups) > 0 {
				if grouped {
					// Show groups separately with headers between them
					return displayTodoListWithGroups(todoList, groups, groupedTodos, ui.OutputFormatTable, showAll)
				} else {
					// Show all todos in single table with GROUP column
					return displayTodoListGitHubStyle(todoList, groups, groupedTodos, showAll)
				}
			}
			return displayTodoListGitHubStyle(todoList, nil, map[string][]api.Todo{"": todos}, showAll)
		},
	}

	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID (overrides default)")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID (overrides default)")
	cmd.Flags().StringVarP(&formatStr, "format", "f", "table", "Output format: table, json, or csv")
	cmd.Flags().StringVar(&jsonFields, "json", "", "Output JSON with specified fields")
	cmd.Flags().BoolVarP(&webView, "web", "w", false, "Open in web browser")
	cmd.Flags().BoolVarP(&showAll, "all", "A", false, "Show all todos including completed ones")
	cmd.Flags().BoolVar(&grouped, "grouped", false, "Show todo groups/sections separately with headers (for organized todo lists)")

	return cmd
}

func outputTodoListJSON(todoList *api.TodoList, todos []api.Todo, _ string) error {
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

	// TODO: If specific fields requested, filter the output
	// Currently, all fields are returned regardless of the fields parameter

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

func displayTodoListWithGroups(todoList *api.TodoList, groups []api.TodoGroup, groupedTodos map[string][]api.Todo, format ui.OutputFormat, showAll bool) error {
	// First, count total todos before any filtering
	totalTodos := 0
	completedTodos := 0
	for _, todos := range groupedTodos {
		for _, todo := range todos {
			totalTodos++
			if todo.Completed {
				completedTodos++
			}
		}
	}

	// Filter todos if not showing all
	if !showAll {
		filteredGroupedTodos := make(map[string][]api.Todo)
		for groupID, todos := range groupedTodos {
			var filteredTodos []api.Todo
			for _, todo := range todos {
				if !todo.Completed {
					filteredTodos = append(filteredTodos, todo)
				}
			}
			filteredGroupedTodos[groupID] = filteredTodos
		}
		groupedTodos = filteredGroupedTodos
	}

	// Terminal display with nice formatting
	if !ui.IsTerminal(os.Stdout) || format == ui.OutputFormatCSV {
		// Non-TTY or CSV format - simple output
		return displayTodoListWithGroupsSimple(todoList, groups, groupedTodos, format)
	}

	// Pretty terminal display
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99"))

	metaStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	// Display title
	fmt.Println(titleStyle.Render(todoList.Title))

	// Display metadata - use the API's completed ratio if available
	var meta string
	if todoList.CompletedRatio != "" {
		meta = todoList.CompletedRatio + " completed"
	} else {
		// Fallback to counting
		totalCompleted := 0
		totalTodos := 0
		for _, group := range groups {
			if todos, ok := groupedTodos[fmt.Sprintf("%d", group.ID)]; ok {
				for _, todo := range todos {
					totalTodos++
					if todo.Completed {
						totalCompleted++
					}
				}
			}
		}
		meta = fmt.Sprintf("%d/%d completed", totalCompleted, totalTodos)
	}

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

	groupTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("75"))

	// Display groups and their todos separately
	for i, group := range groups {
		// Add spacing between groups
		if i > 0 {
			fmt.Println()
		}

		// Display group title with completion ratio
		groupTitle := group.Title
		if group.CompletedRatio != "" {
			groupTitle += " " + metaStyle.Render("("+group.CompletedRatio+")")
		}
		fmt.Println(groupTitleStyle.Render(groupTitle))

		// Get todos for this group
		if todos, ok := groupedTodos[fmt.Sprintf("%d", group.ID)]; ok && len(todos) > 0 {
			// Create GitHub CLI-style table for this group
			table := tableprinter.New(os.Stdout)

			// Add headers dynamically based on TTY mode
			if table.IsTTY() {
				table.AddHeader("", "TODO", "ASSIGNEE", "DUE")
			} else {
				table.AddHeader("STATUS", "TODO", "ASSIGNEE", "STATE", "DUE")
			}

			cs := table.GetColorScheme()

			// Add todos from this group
			for _, todo := range todos {
				// Status column - symbol for TTY, text for non-TTY
				if table.IsTTY() {
					table.AddStatusField(todo.Completed)
				} else {
					if todo.Completed {
						table.AddField("completed")
					} else {
						table.AddField("incomplete")
					}
				}

				// Todo title with completion styling
				title := todo.Content
				if title == "" {
					title = todo.Title
				}
				table.AddTodoField(title, todo.Completed)

				// Get assignees
				assignee := ""
				if len(todo.Assignees) > 0 {
					names := []string{}
					for _, a := range todo.Assignees {
						names = append(names, a.Name)
					}
					assignee = strings.Join(names, ", ")
				}
				table.AddField(assignee, cs.Muted)

				// Add STATE column only for non-TTY
				if !table.IsTTY() {
					if todo.Completed {
						table.AddField("completed")
					} else {
						table.AddField("incomplete")
					}
				}

				// Due date
				due := ""
				if todo.DueOn != nil && *todo.DueOn != "" {
					if dueTime, err := time.Parse("2006-01-02", *todo.DueOn); err == nil {
						due = dueTime.Format("Jan 2")
					}
				}
				table.AddField(due, cs.Muted)

				table.EndRow()
			}

			_ = table.Render()
		} else {
			fmt.Println(metaStyle.Render("  No todos in this group"))
		}
	}

	return nil
}

func displayTodoListWithGroupsSimple(todoList *api.TodoList, groups []api.TodoGroup, groupedTodos map[string][]api.Todo, format ui.OutputFormat) error {
	// Simple output for non-TTY and CSV format
	fmt.Printf("Todo List: %s\n", todoList.Title)
	fmt.Printf("ID: %d\n", todoList.ID)

	totalCompleted := 0
	totalTodos := 0
	for _, group := range groups {
		if todos, ok := groupedTodos[fmt.Sprintf("%d", group.ID)]; ok {
			for _, todo := range todos {
				totalTodos++
				if todo.Completed {
					totalCompleted++
				}
			}
		}
	}
	fmt.Printf("Progress: %d/%d completed\n\n", totalCompleted, totalTodos)

	// Output groups and todos in the requested format
	if format == ui.OutputFormatCSV {
		// Use proper CSV writer for CSV format
		writer := csv.NewWriter(os.Stdout)
		defer writer.Flush()

		// Write header
		if err := writer.Write([]string{"Group", "Status", "Todo", "Due"}); err != nil {
			return fmt.Errorf("failed to write CSV header: %w", err)
		}

		// Write data rows
		for _, group := range groups {
			if todos, ok := groupedTodos[fmt.Sprintf("%d", group.ID)]; ok {
				for _, todo := range todos {
					status := "[ ]"
					if todo.Completed {
						status = "[x]"
					}

					due := ""
					if todo.DueOn != nil {
						due = *todo.DueOn
					}

					record := []string{group.Title, status, todo.Title, due}
					if err := writer.Write(record); err != nil {
						return fmt.Errorf("failed to write CSV record: %w", err)
					}
				}
			}
		}
	} else {
		// Tab-separated output for non-TTY (backwards compatibility)
		fmt.Println("Group\tStatus\tTodo\tDue")
		for _, group := range groups {
			if todos, ok := groupedTodos[fmt.Sprintf("%d", group.ID)]; ok {
				for _, todo := range todos {
					status := "[ ]"
					if todo.Completed {
						status = "[x]"
					}

					due := ""
					if todo.DueOn != nil {
						due = *todo.DueOn
					}

					fmt.Printf("%s\t%s\t%s\t%s\n", group.Title, status, todo.Title, due)
				}
			}
		}
	}

	return nil
}

func outputTodoListWithGroupsJSON(todoList *api.TodoList, groups []api.TodoGroup, groupedTodos map[string][]api.Todo, _ string) error {
	// Combine todo list, groups, and todos data
	groupData := make([]map[string]interface{}, len(groups))
	for i, group := range groups {
		todos := groupedTodos[fmt.Sprintf("%d", group.ID)]
		groupData[i] = map[string]interface{}{
			"id":              group.ID,
			"title":           group.Title,
			"description":     group.Description,
			"completed_ratio": group.CompletedRatio,
			"todos":           todos,
		}
	}

	data := map[string]interface{}{
		"id":          todoList.ID,
		"title":       todoList.Title,
		"description": todoList.Description,
		"created_at":  todoList.CreatedAt,
		"updated_at":  todoList.UpdatedAt,
		"groups":      groupData,
	}

	// TODO: If specific fields requested, filter the output
	// Currently, all fields are returned regardless of the fields parameter

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func displayTodoListGitHubStyle(todoList *api.TodoList, groups []api.TodoGroup, groupedTodos map[string][]api.Todo, showAll bool) error {
	// First, count total todos before any filtering
	totalTodos := 0
	completedTodos := 0
	for _, todos := range groupedTodos {
		for _, todo := range todos {
			totalTodos++
			if todo.Completed {
				completedTodos++
			}
		}
	}

	// Filter todos if not showing all
	filteredGroupedTodos := make(map[string][]api.Todo)
	displayedTodos := 0
	allTodos := []api.Todo{}

	for groupID, todos := range groupedTodos {
		var filteredTodos []api.Todo
		for _, todo := range todos {
			if showAll || !todo.Completed {
				filteredTodos = append(filteredTodos, todo)
				displayedTodos++
				allTodos = append(allTodos, todo)
			}
		}
		filteredGroupedTodos[groupID] = filteredTodos
	}
	groupedTodos = filteredGroupedTodos

	// Display GitHub CLI style summary line
	if showAll {
		fmt.Printf("Showing %d of %d todos in %s\n\n", displayedTodos, totalTodos, todoList.Title)
	} else {
		fmt.Printf("Showing %d of %d open todos in %s\n\n", displayedTodos, totalTodos, todoList.Title)
	}

	// Create GitHub CLI-style table
	table := tableprinter.New(os.Stdout)

	// Add headers dynamically based on TTY mode and groups
	if table.IsTTY() {
		if len(groups) > 0 {
			table.AddHeader("ID", "", "TODO", "GROUP", "ASSIGNEE", "DUE")
		} else {
			table.AddHeader("ID", "", "TODO", "ASSIGNEE", "DUE")
		}
	} else {
		// Add STATE column for non-TTY mode (machine readable)
		if len(groups) > 0 {
			table.AddHeader("ID", "STATUS", "TODO", "GROUP", "ASSIGNEE", "STATE", "DUE")
		} else {
			table.AddHeader("ID", "STATUS", "TODO", "ASSIGNEE", "STATE", "DUE")
		}
	}

	cs := table.GetColorScheme()

	// Add all todos to single table
	if len(groups) > 0 {
		// With groups
		for _, group := range groups {
			if todos, ok := groupedTodos[fmt.Sprintf("%d", group.ID)]; ok {
				for _, todo := range todos {
					// ID column
					table.AddField(fmt.Sprintf("%d", todo.ID))

					// Status column - symbol for TTY, text for non-TTY
					if table.IsTTY() {
						table.AddStatusField(todo.Completed)
					} else {
						if todo.Completed {
							table.AddField("completed")
						} else {
							table.AddField("incomplete")
						}
					}

					// Todo title with completion styling
					title := todo.Content
					if title == "" {
						title = todo.Title
					}
					table.AddTodoField(title, todo.Completed)

					// Group name with cyan color (like GitHub CLI branch names)
					table.AddField(group.Title, cs.Cyan)

					// Get assignees
					assignee := ""
					if len(todo.Assignees) > 0 {
						names := []string{}
						for _, a := range todo.Assignees {
							names = append(names, a.Name)
						}
						assignee = strings.Join(names, ", ")
					}
					table.AddField(assignee, cs.Muted)

					// Add STATE column only for non-TTY
					if !table.IsTTY() {
						if todo.Completed {
							table.AddField("completed")
						} else {
							table.AddField("incomplete")
						}
					}

					// Due date
					due := ""
					if todo.DueOn != nil && *todo.DueOn != "" {
						if dueTime, err := time.Parse("2006-01-02", *todo.DueOn); err == nil {
							due = dueTime.Format("Jan 2")
						}
					}
					table.AddField(due, cs.Muted)

					table.EndRow()
				}
			}
		}
	} else {
		// Without groups
		for _, todo := range allTodos {
			// ID column
			table.AddField(fmt.Sprintf("%d", todo.ID))

			// Status column - symbol for TTY, text for non-TTY
			if table.IsTTY() {
				table.AddStatusField(todo.Completed)
			} else {
				if todo.Completed {
					table.AddField("completed")
				} else {
					table.AddField("incomplete")
				}
			}

			// Todo title with completion styling
			title := todo.Content
			if title == "" {
				title = todo.Title
			}
			table.AddTodoField(title, todo.Completed)

			// Get assignees
			assignee := ""
			if len(todo.Assignees) > 0 {
				names := []string{}
				for _, a := range todo.Assignees {
					names = append(names, a.Name)
				}
				assignee = strings.Join(names, ", ")
			}
			table.AddField(assignee, cs.Muted)

			// Add STATE column only for non-TTY
			if !table.IsTTY() {
				if todo.Completed {
					table.AddField("completed")
				} else {
					table.AddField("incomplete")
				}
			}

			// Due date
			due := ""
			if todo.DueOn != nil && *todo.DueOn != "" {
				if dueTime, err := time.Parse("2006-01-02", *todo.DueOn); err == nil {
					due = dueTime.Format("Jan 2")
				}
			}
			table.AddField(due, cs.Muted)

			table.EndRow()
		}
	}

	return table.Render()
}
