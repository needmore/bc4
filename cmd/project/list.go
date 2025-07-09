package project

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

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all projects",
		Long:  `List all projects in your Basecamp account. Use 'project select' for interactive selection.`,
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

			// Get token
			token, err := authClient.GetToken(accountID)
			if err != nil {
				return fmt.Errorf("failed to get auth token: %w", err)
			}

			// Create API client
			apiClient := api.NewClient(accountID, token.AccessToken)

			// Fetch projects
			projects, err := apiClient.GetProjects(context.Background())
			if err != nil {
				return fmt.Errorf("failed to fetch projects: %w", err)
			}

			// Sort projects alphabetically
			sortProjectsByName(projects)

			// Get default project ID
			defaultProjectID := cfg.DefaultProject
			if defaultProjectID == "" && cfg.Accounts != nil {
				if acc, ok := cfg.Accounts[accountID]; ok {
					defaultProjectID = acc.DefaultProject
				}
			}

			// Output JSON if requested
			if jsonOutput {
				return outputJSON(projects)
			}

			// Check if there are any projects
			if len(projects) == 0 {
				fmt.Println("No projects found.")
				return nil
			}

			// Get terminal width
			termWidth := ui.GetTerminalWidth()

			// Calculate column widths based on terminal width
			// Minimum widths: Name=20, ID=8, Description=20
			nameWidth := 40
			idWidth := 10
			descWidth := 50
			
			// Total minimum width needed
			minWidth := nameWidth + idWidth + descWidth + 6 // 6 for separators
			
			if termWidth < minWidth {
				// Shrink columns proportionally
				nameWidth = (termWidth * 40) / minWidth
				idWidth = (termWidth * 10) / minWidth
				descWidth = termWidth - nameWidth - idWidth - 6
				
				// Ensure minimum widths
				if nameWidth < 20 {
					nameWidth = 20
				}
				if idWidth < 8 {
					idWidth = 8
				}
				if descWidth < 10 {
					descWidth = 10
				}
			} else if termWidth > minWidth {
				// Give extra space to description column
				descWidth = termWidth - nameWidth - idWidth - 6
			}

			// Create table
			columns := []table.Column{
				{Title: "", Width: nameWidth},
				{Title: "", Width: idWidth},
				{Title: "", Width: descWidth},
			}

			rows := []table.Row{}
			defaultIndex := 0
			for i, project := range projects {
				idStr := strconv.FormatInt(project.ID, 10)
				name := ui.TruncateString(project.Name, nameWidth-3)
				desc := ui.TruncateString(project.Description, descWidth-3)
				
				// Track which row is the default
				if idStr == defaultProjectID {
					defaultIndex = i
				}
				
				rows = append(rows, table.Row{
					name,
					idStr,
					desc,
				})
			}

			t := table.New(
				table.WithColumns(columns),
				table.WithRows(rows),
				table.WithHeight(len(rows)+2), // Include space for header and borders
			)

			// Style the table with subtle list highlighting
			t = ui.StyleTableForList(t)
			
			// Set cursor to the default project
			t.SetCursor(defaultIndex)

			// Print the table, skipping the empty header row
			tableView := t.View()
			lines := strings.Split(tableView, "\n")
			if len(lines) > 2 {
				// Skip first line (top border) and second line (empty header)
				// Keep the top border but skip the header
				result := lines[0] + "\n" + strings.Join(lines[2:], "\n")
				fmt.Println(ui.BaseTableStyle.Render(result))
			} else {
				fmt.Println(ui.BaseTableStyle.Render(tableView))
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID (overrides default)")

	return cmd
}

func sortProjectsByName(projects []api.Project) {
	// Using bubble sort for simplicity, could use sort.Slice
	for i := 0; i < len(projects); i++ {
		for j := i + 1; j < len(projects); j++ {
			if strings.ToLower(projects[i].Name) > strings.ToLower(projects[j].Name) {
				projects[i], projects[j] = projects[j], projects[i]
			}
		}
	}
}

func outputJSON(projects []api.Project) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(projects)
}