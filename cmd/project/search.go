package project

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
	"github.com/needmore/bc4/internal/ui/tableprinter"
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
			apiClient := api.NewModularClient(accountID, token.AccessToken)
			projectOps := apiClient.Projects()

			// Fetch all projects
			allProjects, err := projectOps.GetProjects(context.Background())
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

			// Print results count
			fmt.Printf("\nFound %d project%s matching \"%s\":\n\n",
				len(matchingProjects),
				pluralize(len(matchingProjects)),
				strings.Join(args, " "))

			// Create GitHub CLI-style table
			table := tableprinter.New(os.Stdout)

			// Add headers dynamically based on TTY mode
			if table.IsTTY() {
				table.AddHeader("NAME", "ID", "DESCRIPTION", "UPDATED")
			} else {
				// Add STATE column for non-TTY mode (machine readable)
				table.AddHeader("NAME", "ID", "DESCRIPTION", "STATE", "UPDATED")
			}

			// Get color scheme for consistent styling
			cs := table.GetColorScheme()

			// Add projects to table
			for _, project := range matchingProjects {
				// Add project name with appropriate coloring
				state := "active"
				table.AddProjectField(project.Name, state)

				// Add ID field
				table.AddIDField(fmt.Sprintf("%d", project.ID), state)

				// Add description with truncation
				desc := project.Description
				if len(desc) > 50 {
					desc = desc[:47] + "..."
				}
				table.AddField(desc, cs.Muted)

				// Add STATE column only for non-TTY
				if !table.IsTTY() {
					table.AddField(state)
				}

				// Add updated time
				table.AddField(project.UpdatedAt, cs.Muted)

				table.EndRow()
			}

			return table.Render()
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
