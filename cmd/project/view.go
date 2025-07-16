package project

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/needmore/bc4/internal/ui"
	"github.com/needmore/bc4/internal/utils"
)

func newViewCmd(f *factory.Factory) *cobra.Command {
	var jsonOutput bool
	var accountID string
	var noPager bool

	cmd := &cobra.Command{
		Use:   "view [project-id or URL]",
		Short: "View project details",
		Long: `View detailed information about a specific project.

You can specify the project using either:
- A numeric ID (e.g., "12345")
- A project name (partial match supported)
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/projects/12345")`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get config and auth through factory
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			authClient, err := f.AuthClient()
			if err != nil {
				return err
			}

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
				// Check if it's a URL
				if parser.IsBasecampURL(args[0]) {
					parsedURL, err := parser.ParseBasecampURL(args[0])
					if err != nil {
						return fmt.Errorf("invalid Basecamp URL: %s", args[0])
					}
					if parsedURL.ResourceType != parser.ResourceTypeProject {
						return fmt.Errorf("URL is not for a project: %s", args[0])
					}
					projectID = strconv.FormatInt(parsedURL.ResourceID, 10)
					// Override account ID if provided in URL
					if parsedURL.AccountID > 0 {
						accountID = strconv.FormatInt(parsedURL.AccountID, 10)
					}
				} else {
					projectID = args[0]
				}
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

			// Create API client through factory
			// If accountID was specified, use a new factory with that account
			if accountID != "" {
				f = f.WithAccount(accountID)
			}
			apiClient, err := f.ApiClient()
			if err != nil {
				return err
			}
			projectOps := apiClient.Projects()

			// Try to fetch as ID first
			project, err = projectOps.GetProject(f.Context(), projectID)
			if err != nil {
				// If that fails and the input doesn't look like a number, try searching by name
				if _, parseErr := strconv.ParseInt(projectID, 10, 64); parseErr != nil {
					// Search for projects matching the name
					allProjects, fetchErr := projectOps.GetProjects(f.Context())
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

			// Prepare output for pager
			var buf bytes.Buffer

			// Display project details
			fmt.Fprintln(&buf)
			fmt.Fprintln(&buf, ui.TitleStyle.Render(project.Name))
			fmt.Fprintln(&buf)

			if project.Description != "" {
				fmt.Fprintf(&buf, "%s %s\n", ui.LabelStyle.Render("Description:"), ui.ValueStyle.Render(project.Description))
			}

			fmt.Fprintf(&buf, "%s %s\n", ui.LabelStyle.Render("ID:"), ui.ValueStyle.Render(strconv.FormatInt(project.ID, 10)))

			// Only show purpose if it's not empty and not "topic"
			if project.Purpose != "" && project.Purpose != "topic" {
				fmt.Fprintf(&buf, "%s %s\n", ui.LabelStyle.Render("Purpose:"), ui.ValueStyle.Render(project.Purpose))
			}

			// Parse and format dates
			if created, err := time.Parse(time.RFC3339, project.CreatedAt); err == nil {
				fmt.Fprintf(&buf, "%s %s\n", ui.LabelStyle.Render("Created:"), ui.ValueStyle.Render(created.Format("January 2, 2006")))
			}

			if updated, err := time.Parse(time.RFC3339, project.UpdatedAt); err == nil {
				fmt.Fprintf(&buf, "%s %s\n", ui.LabelStyle.Render("Updated:"), ui.ValueStyle.Render(updated.Format("January 2, 2006")))
			}

			fmt.Fprintln(&buf)

			// Display using pager
			pagerOpts := &utils.PagerOptions{
				Pager:   cfg.Preferences.Pager,
				NoPager: noPager,
			}
			return utils.ShowInPager(buf.String(), pagerOpts)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID (overrides default)")
	cmd.Flags().BoolVar(&noPager, "no-pager", false, "Disable pager for output")

	return cmd
}
