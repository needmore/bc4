package people

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/ui"
)

func newListCmd(f *factory.Factory) *cobra.Command {
	var jsonOutput bool
	var formatStr string
	var projectID string
	var accountID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all people",
		Long: `List all people in your Basecamp account, or filter by project.

By default, lists all people visible to you in the account.
Use --project to list only people with access to a specific project.`,
		Aliases: []string{"ls"},
		Example: `  # List all people in the account
  bc4 people list

  # List people in a specific project
  bc4 people list --project 12345

  # Output as JSON
  bc4 people list --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Apply overrides if specified
			f = f.ApplyOverrides(accountID, projectID)

			// Get API client from factory
			client, err := f.ApiClient()
			if err != nil {
				return err
			}
			peopleOps := client.People()

			// Fetch people based on whether project is specified
			var people []api.Person
			var listContext string

			if projectID != "" {
				// List people in specific project
				resolvedProjectID, err := f.ProjectID()
				if err != nil {
					return err
				}
				people, err = peopleOps.GetProjectPeople(f.Context(), resolvedProjectID)
				if err != nil {
					return fmt.Errorf("failed to fetch project people: %w", err)
				}
				listContext = fmt.Sprintf("project %s", resolvedProjectID)
			} else {
				// List all people in account
				people, err = peopleOps.GetAllPeople(f.Context())
				if err != nil {
					return fmt.Errorf("failed to fetch people: %w", err)
				}
				listContext = "account"
			}

			// Sort people alphabetically by name
			sort.Slice(people, func(i, j int) bool {
				return strings.ToLower(people[i].Name) < strings.ToLower(people[j].Name)
			})

			// Parse output format
			format, err := ui.ParseOutputFormat(formatStr)
			if err != nil {
				return err
			}

			// Handle legacy JSON flag
			if jsonOutput {
				format = ui.OutputFormatJSON
			}

			// Handle JSON output
			if format == ui.OutputFormatJSON {
				return outputPeopleJSON(people)
			}

			// Check if there are any people
			if len(people) == 0 {
				fmt.Printf("No people found in %s.\n", listContext)
				return nil
			}

			// Render the people table with role column
			return renderPeopleTable(people, true)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON (deprecated, use --format=json)")
	cmd.Flags().StringVarP(&formatStr, "format", "f", "table", "Output format: table, json, or csv")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Filter by project ID")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID (overrides default)")

	return cmd
}

func outputPeopleJSON(people []api.Person) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(people)
}
