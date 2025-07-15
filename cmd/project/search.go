package project

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/ui/tableprinter"
)

func newSearchCmd(f *factory.Factory) *cobra.Command {
	var jsonOutput bool
	var accountID string

	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search projects by name",
		Long:  `Search for projects whose names contain the specified query string (case-insensitive).`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.ToLower(strings.Join(args, " "))

			// Get auth through factory
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

			// Fetch all projects
			allProjects, err := projectOps.GetProjects(f.Context())
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
