package todo

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/ui"
	"github.com/needmore/bc4/internal/ui/tableprinter"
)

func newListsCmd(f *factory.Factory) *cobra.Command {
	var jsonOutput bool
	var accountID string
	var projectID string
	var formatStr string

	cmd := &cobra.Command{
		Use:     "lists",
		Short:   "List all todo lists in a project",
		Long:    `List all todo lists in the current or specified project.`,
		Aliases: []string{"ls"},
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

			// Get the todo set for the project
			todoSet, err := todoOps.GetProjectTodoSet(f.Context(), resolvedProjectID)
			if err != nil {
				return fmt.Errorf("failed to get todo set: %w", err)
			}

			// Fetch todo lists
			todoLists, err := todoOps.GetTodoLists(f.Context(), resolvedProjectID, todoSet.ID)
			if err != nil {
				return fmt.Errorf("failed to fetch todo lists: %w", err)
			}

			// Sort todo lists alphabetically
			sortTodoListsByName(todoLists)

			// Get default todo list ID from config
			defaultTodoListID := ""
			if cfg.Accounts != nil && cfg.Accounts[resolvedAccountID].ProjectDefaults != nil {
				if projDefaults, ok := cfg.Accounts[resolvedAccountID].ProjectDefaults[resolvedProjectID]; ok {
					defaultTodoListID = projDefaults.DefaultTodoList
				}
			}

			// Check if there are any todo lists
			if len(todoLists) == 0 {
				fmt.Println("No todo lists found in this project.")
				return nil
			}

			// Parse output format
			format, err := ui.ParseOutputFormat(formatStr)
			if err != nil {
				return err
			}

			// Handle legacy JSON flag
			if jsonOutput {
				format = ui.OutputFormatJSON
			}

			// Handle JSON output directly
			if format == ui.OutputFormatJSON {
				// TODO: Implement JSON output for todo lists
				return fmt.Errorf("JSON output not yet implemented for todo lists")
			}

			// Create new GitHub CLI-style table
			table := tableprinter.New(os.Stdout)

			// Add headers dynamically based on TTY mode (like GitHub CLI)
			if table.IsTTY() {
				table.AddHeader("ID", "NAME", "PROGRESS", "UPDATED")
			} else {
				// Add STATE column for non-TTY mode (machine readable)
				table.AddHeader("ID", "NAME", "PROGRESS", "STATE", "UPDATED")
			}

			// Add todo lists to table
			for _, todoList := range todoLists {
				idStr := strconv.FormatInt(todoList.ID, 10)

				// Add ID field with default indicator
				if idStr == defaultTodoListID {
					if table.IsTTY() {
						table.AddIDField(idStr+"*", "active") // Add asterisk for default
					} else {
						table.AddIDField(idStr, "active")
					}
				} else {
					table.AddIDField(idStr, "active")
				}

				// Add todo list name with cyan color (like GitHub CLI branch names)
				cs := table.GetColorScheme()
				table.AddField(todoList.Title, cs.TodoListName)

				// Add progress/completion ratio
				status := todoList.CompletedRatio
				if status == "" {
					status = "0/0"
				}
				table.AddField(status, cs.Muted)

				// Add STATE column only for non-TTY
				if !table.IsTTY() {
					table.AddField("active")
				}

				// Add updated time - use UpdatedAt if available
				timeStr := todoList.UpdatedAt
				if timeStr == "" {
					timeStr = "N/A"
				}
				table.AddField(timeStr, cs.Muted)

				table.EndRow()
			}

			return table.Render()
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON (deprecated, use --format=json)")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID (overrides default)")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID (overrides default)")
	cmd.Flags().StringVarP(&formatStr, "format", "f", "table", "Output format: table, json, or csv")

	return cmd
}

func sortTodoListsByName(todoLists []api.TodoList) {
	// Using bubble sort for simplicity
	for i := 0; i < len(todoLists); i++ {
		for j := i + 1; j < len(todoLists); j++ {
			if strings.ToLower(todoLists[i].Title) > strings.ToLower(todoLists[j].Title) {
				todoLists[i], todoLists[j] = todoLists[j], todoLists[i]
			}
		}
	}
}
