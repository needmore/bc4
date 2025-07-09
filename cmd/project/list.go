package project

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	
	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
)

// Styles for the table
var (
	baseStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))
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
			termWidth := 100 // default
			if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && width > 0 {
				termWidth = width - 4 // Leave some margin for borders
			}

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
				{Title: "Name", Width: nameWidth},
				{Title: "ID", Width: idWidth},
				{Title: "Description", Width: descWidth},
			}

			rows := []table.Row{}
			for _, project := range projects {
				name := project.Name
				if len(name) > nameWidth-3 {
					name = name[:nameWidth-6] + "..."
				}
				
				desc := project.Description
				// Don't fall back to Purpose - just leave blank if no description
				// Truncate description if too long
				if len(desc) > descWidth-3 {
					desc = desc[:descWidth-6] + "..."
				}
				
				rows = append(rows, table.Row{
					name,
					strconv.FormatInt(project.ID, 10),
					desc,
				})
			}

			t := table.New(
				table.WithColumns(columns),
				table.WithRows(rows),
				table.WithHeight(len(rows)+2),
			)

			// Style the table
			s := table.DefaultStyles()
			s.Header = s.Header.
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				BorderBottom(true).
				Bold(false)
			t.SetStyles(s)

			// Just render and print the table
			fmt.Println(baseStyle.Render(t.View()))
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