package todo

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/spf13/cobra"
	
	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
	"github.com/needmore/bc4/internal/ui"
)

func newListCmd() *cobra.Command {
	var jsonOutput bool
	var accountID string
	var projectID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all todo lists in a project",
		Long:  `List all todo lists in the current or specified project.`,
		Aliases: []string{"ls"},
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
				return fmt.Errorf("no project specified and no default project set. Use 'bc4 project select' to set a default project")
			}

			// Get token
			token, err := authClient.GetToken(accountID)
			if err != nil {
				return fmt.Errorf("failed to get auth token: %w", err)
			}

			// Create API client
			apiClient := api.NewClient(accountID, token.AccessToken)

			// Get the todo set for the project
			todoSet, err := apiClient.GetProjectTodoSet(context.Background(), projectID)
			if err != nil {
				return fmt.Errorf("failed to get todo set: %w", err)
			}

			// Fetch todo lists
			todoLists, err := apiClient.GetTodoLists(context.Background(), projectID, todoSet.ID)
			if err != nil {
				return fmt.Errorf("failed to fetch todo lists: %w", err)
			}

			// Sort todo lists alphabetically
			sortTodoListsByName(todoLists)

			// Get default todo list ID from config
			defaultTodoListID := ""
			if cfg.Accounts != nil && cfg.Accounts[accountID].ProjectDefaults != nil {
				if projDefaults, ok := cfg.Accounts[accountID].ProjectDefaults[projectID]; ok {
					defaultTodoListID = projDefaults.DefaultTodoList
				}
			}

			// Output JSON if requested
			if jsonOutput {
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(todoLists)
			}

			// Check if there are any todo lists
			if len(todoLists) == 0 {
				fmt.Println("No todo lists found in this project.")
				return nil
			}
			

			// Get terminal width
			termWidth := ui.GetTerminalWidth()

			// Calculate column widths based on terminal width
			nameWidth := 50
			idWidth := 12
			statusWidth := 10
			
			// Total minimum width needed
			minWidth := nameWidth + idWidth + statusWidth + 6 // 6 for separators
			
			if termWidth < minWidth {
				// Shrink columns proportionally
				nameWidth = (termWidth * 50) / minWidth
				statusWidth = (termWidth * 10) / minWidth
				
				// Ensure minimum widths
				if nameWidth < 20 {
					nameWidth = 20
				}
				if statusWidth < 8 {
					statusWidth = 8
				}
			} else if termWidth > minWidth {
				// Give extra space to name column
				nameWidth = termWidth - idWidth - statusWidth - 6
			}

			// Create table
			columns := []table.Column{
				{Title: "", Width: nameWidth},
				{Title: "", Width: idWidth},
				{Title: "", Width: statusWidth},
			}

			rows := []table.Row{}
			defaultIndex := 0
			for i, todoList := range todoLists {
				idStr := strconv.FormatInt(todoList.ID, 10)
				name := ui.TruncateString(todoList.Title, nameWidth-3)
				status := todoList.CompletedRatio
				if status == "" {
					status = "0/0"
				}
				
				// Track which row is the default
				if idStr == defaultTodoListID {
					defaultIndex = i
				}
				
				rows = append(rows, table.Row{
					name,
					idStr,
					status,
				})
			}

			// Calculate the proper height - we need all rows plus borders
			// Since we're skipping the header, we need to ensure all data rows are visible
			tableHeight := len(rows) + 1  // +1 for borders
			
			t := table.New(
				table.WithColumns(columns),
				table.WithRows(rows),
				table.WithHeight(tableHeight),
			)

			// Style the table with subtle list highlighting
			t = ui.StyleTableForList(t)
			
			// Set cursor to the default todo list (for highlighting)
			t.SetCursor(defaultIndex)
			
			// Make sure table shows all rows by blurring focus
			t.Blur()

			// Print the table, skipping the empty header row
			tableView := t.View()
			lines := strings.Split(tableView, "\n")
			
			if len(lines) > 1 {
				// Skip the first line (empty header), keep all data rows
				result := strings.Join(lines[1:], "\n")
				fmt.Println(ui.BaseTableStyle.Render(result))
			} else {
				fmt.Println(ui.BaseTableStyle.Render(tableView))
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID (overrides default)")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID (overrides default)")

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