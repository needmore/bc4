package project

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/needmore/bc4/internal/ui"
	"github.com/spf13/cobra"
)

func newEditCmd(f *factory.Factory) *cobra.Command {
	var name string
	var description string
	var accountID string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "edit <project-id|url>",
		Short: "Update project details",
		Long: `Update the name or description of a Basecamp project.

You can specify the project using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/projects/89012345")

If no flags are provided, interactive mode will be used with current values pre-filled.

To clear the description, use --description with an empty string: --description ""

Examples:
  bc4 project edit 12345 --name "New Name"
  bc4 project edit 12345 --description "New description"
  bc4 project edit 12345 --description ""       # Clear the description
  bc4 project edit 12345                        # Interactive mode`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var projectID string

			// Parse the argument - could be a URL or ID
			if parser.IsBasecampURL(args[0]) {
				parsed, err := parser.ParseBasecampURL(args[0])
				if err != nil {
					return fmt.Errorf("invalid Basecamp URL: %w", err)
				}
				if parsed.ResourceType != parser.ResourceTypeProject {
					return fmt.Errorf("URL is not a project URL: %s", args[0])
				}
				projectID = strconv.FormatInt(parsed.ResourceID, 10)
				if accountID == "" {
					accountID = strconv.FormatInt(parsed.AccountID, 10)
				}
			} else {
				projectID = args[0]
			}

			// Apply overrides if specified
			if accountID != "" {
				f = f.WithAccount(accountID)
			}

			// Get API client from factory
			client, err := f.ApiClient()
			if err != nil {
				return err
			}

			// Fetch current project
			project, err := client.Projects().GetProject(f.Context(), projectID)
			if err != nil {
				return fmt.Errorf("failed to fetch project: %w", err)
			}

			// Interactive mode if no flags provided
			if name == "" && description == "" {
				// Pre-fill with current values
				name = project.Name
				description = project.Description

				if err := huh.NewInput().
					Title("Project Name").
					Description("Current: " + project.Name).
					Value(&name).
					Validate(func(s string) error {
						if s == "" {
							return fmt.Errorf("project name is required")
						}
						return nil
					}).
					Run(); err != nil {
					return err
				}

				if err := huh.NewText().
					Title("Description").
					Description("Current: " + project.Description).
					Value(&description).
					Run(); err != nil {
					return err
				}

				// Check if anything changed
				if name == project.Name && description == project.Description {
					fmt.Println("No changes made")
					return nil
				}
			}

			// Build update request
			req := api.ProjectUpdateRequest{}
			if name != "" {
				req.Name = name
			}
			if description != "" || cmd.Flags().Changed("description") {
				req.Description = description
			}

			// Update the project
			updatedProject, err := client.Projects().UpdateProject(f.Context(), projectID, req)
			if err != nil {
				return fmt.Errorf("failed to update project: %w", err)
			}

			// Output
			if jsonOutput {
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(updatedProject)
			}

			if ui.IsTerminal(os.Stdout) {
				fmt.Printf("✓ Updated project: %s (#%d)\n", updatedProject.Name, updatedProject.ID)
			} else {
				fmt.Printf("%d\n", updatedProject.ID)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "New project name")
	cmd.Flags().StringVarP(&description, "description", "d", "", "New project description")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}
