package project

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/needmore/bc4/internal/ui"
	"github.com/spf13/cobra"
)

func newCopyCmd(f *factory.Factory) *cobra.Command {
	var name string
	var description string
	var accountID string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "copy <project-id|url>",
		Short: "Duplicate a project from a template",
		Long: `Create a new project by duplicating an existing project template.

The source project should be a template project. The new project will include
the same structure and tools as the template.

You can specify the source project using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/projects/89012345")

Examples:
  bc4 project copy 12345                       # Interactive mode
  bc4 project copy 12345 --name "New Project"  # Create with specified name
  bc4 project copy 12345 --name "My Project" --description "Project description"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var sourceProjectID string

			// Parse the argument - could be a URL or ID
			if parser.IsBasecampURL(args[0]) {
				parsed, err := parser.ParseBasecampURL(args[0])
				if err != nil {
					return fmt.Errorf("invalid Basecamp URL: %w", err)
				}
				if parsed.ResourceType != parser.ResourceTypeProject {
					return fmt.Errorf("URL is not a project URL: %s", args[0])
				}
				sourceProjectID = strconv.FormatInt(parsed.ResourceID, 10)
				if accountID == "" {
					accountID = strconv.FormatInt(parsed.AccountID, 10)
				}
			} else {
				sourceProjectID = args[0]
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

			// Fetch source project to show info
			sourceProject, err := client.Projects().GetProject(f.Context(), sourceProjectID)
			if err != nil {
				return fmt.Errorf("failed to fetch source project: %w", err)
			}

			// Interactive mode if no name provided
			if name == "" {
				// Default name suggestion
				defaultName := sourceProject.Name + " (Copy)"

				if err := huh.NewInput().
					Title("New Project Name").
					Description(fmt.Sprintf("Creating from template: %s", sourceProject.Name)).
					Value(&name).
					Placeholder(defaultName).
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
					Title("Description (optional)").
					Description("Enter an optional description for the new project").
					Value(&description).
					Run(); err != nil {
					return err
				}
			}

			// Copy the project
			newProject, err := client.Projects().CopyProject(f.Context(), sourceProjectID, name, description)
			if err != nil {
				return fmt.Errorf("failed to copy project: %w", err)
			}

			// Output
			if jsonOutput {
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(newProject)
			}

			if ui.IsTerminal(os.Stdout) {
				fmt.Printf("✓ Created project: %s (#%d) from template: %s\n", newProject.Name, newProject.ID, sourceProject.Name)
			} else {
				fmt.Printf("%d\n", newProject.ID)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Name for the new project")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Description for the new project")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}
