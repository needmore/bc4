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

func newSearchCmd() *cobra.Command {
	var jsonOutput bool
	var accountID string

	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search projects by name",
		Long:  `Search for projects whose names contain the specified query string (case-insensitive).`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.ToLower(strings.Join(args, " "))

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

			// Fetch all projects
			allProjects, err := apiClient.GetProjects(context.Background())
			if err != nil {
				return fmt.Errorf("failed to fetch projects: %w", err)
			}

			// Filter projects by query
			var matchingProjects []api.Project
			for _, project := range allProjects {
				if strings.Contains(strings.ToLower(project.Name), query) {
					matchingProjects = append(matchingProjects, project)
				}
			}

			// Sort matching projects alphabetically
			sortProjectsByName(matchingProjects)

			// Output JSON if requested
			if jsonOutput {
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(matchingProjects)
			}

			// Check if there are any matching projects
			if len(matchingProjects) == 0 {
				fmt.Printf("No projects found matching \"%s\".\n", strings.Join(args, " "))
				return nil
			}

			// Create table
			columns := []table.Column{
				{Title: "", Width: 40},
				{Title: "", Width: 10},
				{Title: "", Width: 50},
			}

			rows := []table.Row{}
			for _, project := range matchingProjects {
				desc := project.Description
				// Truncate description if too long
				if len(desc) > 47 {
					desc = desc[:44] + "..."
				}

				rows = append(rows, table.Row{
					project.Name,
					strconv.FormatInt(project.ID, 10),
					desc,
				})
			}

			t := table.New(
				table.WithColumns(columns),
				table.WithRows(rows),
				table.WithHeight(len(rows)+1),
			)

			// Apply display-only table styling (no row selection)
			t = ui.StyleTableForDisplay(t)

			// Make sure table shows all rows
			t.Blur()

			// Print results count
			fmt.Printf("\nFound %d project%s matching \"%s\":\n\n",
				len(matchingProjects),
				pluralize(len(matchingProjects)),
				strings.Join(args, " "))

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

	return cmd
}

func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
