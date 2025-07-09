package todo

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

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
	var formatStr string

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all todo lists in a project",
		Long:    `List all todo lists in the current or specified project.`,
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

			// Create output config
			config := ui.NewOutputConfig(os.Stdout)
			config.Format = format
			config.NoHeaders = true // We'll add custom formatting

			// Create table writer
			tw := ui.NewTableWriter(config)

			// Add rows
			for _, todoList := range todoLists {
				idStr := strconv.FormatInt(todoList.ID, 10)
				status := todoList.CompletedRatio
				if status == "" {
					status = "0/0"
				}

				row := []string{
					todoList.Title,
					idStr,
					status,
				}

				// Add a marker for the default todo list
				if idStr == defaultTodoListID {
					if config.Format == ui.OutputFormatTable && ui.IsTerminal(os.Stdout) && !config.NoColor {
						// Add a subtle indicator for the default todo list
						row[0] = "→ " + row[0]
					}
				}

				tw.AddRow(row)
			}

			return tw.Render()
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON (deprecated, use --format=json)")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID (overrides default)")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID (overrides default)")
	cmd.Flags().StringVarP(&formatStr, "format", "f", "table", "Output format: table, json, or tsv")

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
