package project

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
	"github.com/needmore/bc4/internal/ui"
)

func newViewCmd() *cobra.Command {
	var jsonOutput bool
	var accountID string

	cmd := &cobra.Command{
		Use:   "view [project-id]",
		Short: "View project details",
		Long:  `View detailed information about a specific project.`,
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

			// Get project ID from args or default
			var projectID string
			var project *api.Project

			if len(args) > 0 {
				projectID = args[0]
			} else {
				// Use default project
				projectID = cfg.DefaultProject
				if projectID == "" && cfg.Accounts != nil {
					if acc, ok := cfg.Accounts[accountID]; ok {
						projectID = acc.DefaultProject
					}
				}
				if projectID == "" {
					return fmt.Errorf("no project specified and no default project set")
				}
			}

			// Get token
			token, err := authClient.GetToken(accountID)
			if err != nil {
				return fmt.Errorf("failed to get auth token: %w", err)
			}

			// Create API client
			apiClient := api.NewClient(accountID, token.AccessToken)

			// Try to fetch as ID first
			project, err = apiClient.GetProject(context.Background(), projectID)
			if err != nil {
				// If that fails and the input doesn't look like a number, try searching by name
				if _, parseErr := strconv.ParseInt(projectID, 10, 64); parseErr != nil {
					// Search for projects matching the name
					allProjects, fetchErr := apiClient.GetProjects(context.Background())
					if fetchErr != nil {
						return fmt.Errorf("failed to fetch projects: %w", fetchErr)
					}

					// Find first matching project (case-insensitive)
					searchTerm := strings.ToLower(projectID)
					for _, p := range allProjects {
						if strings.Contains(strings.ToLower(p.Name), searchTerm) {
							proj := p // Create a copy to get the address
							project = &proj
							break
						}
					}

					if project == nil {
						return fmt.Errorf("no project found matching '%s'", projectID)
					}
				} else {
					return fmt.Errorf("failed to fetch project: %w", err)
				}
			}

			// Output JSON if requested
			if jsonOutput {
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(project)
			}

			// Display project details
			fmt.Println()
			fmt.Println(ui.TitleStyle.Render(project.Name))
			fmt.Println()

			if project.Description != "" {
				fmt.Printf("%s %s\n", ui.LabelStyle.Render("Description:"), ui.ValueStyle.Render(project.Description))
			}

			fmt.Printf("%s %s\n", ui.LabelStyle.Render("ID:"), ui.ValueStyle.Render(strconv.FormatInt(project.ID, 10)))

			// Only show purpose if it's not empty and not "topic"
			if project.Purpose != "" && project.Purpose != "topic" {
				fmt.Printf("%s %s\n", ui.LabelStyle.Render("Purpose:"), ui.ValueStyle.Render(project.Purpose))
			}

			// Parse and format dates
			if created, err := time.Parse(time.RFC3339, project.CreatedAt); err == nil {
				fmt.Printf("%s %s\n", ui.LabelStyle.Render("Created:"), ui.ValueStyle.Render(created.Format("January 2, 2006")))
			}

			if updated, err := time.Parse(time.RFC3339, project.UpdatedAt); err == nil {
				fmt.Printf("%s %s\n", ui.LabelStyle.Render("Updated:"), ui.ValueStyle.Render(updated.Format("January 2, 2006")))
			}

			fmt.Println()

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID (overrides default)")

	return cmd
}
