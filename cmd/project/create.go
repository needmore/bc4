package project

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/ui"
	"github.com/spf13/cobra"
)

func newCreateCmd(f *factory.Factory) *cobra.Command {
	var name string
	var description string
	var accountID string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new project",
		Long: `Create a new Basecamp project.

You can specify the project name and description via flags, or use interactive mode.

Examples:
  bc4 project create                              # Interactive mode
  bc4 project create --name "My Project"          # Create with name only
  bc4 project create --name "My Project" --description "Project description"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Apply overrides if specified
			if accountID != "" {
				f = f.WithAccount(accountID)
			}

			// Get API client from factory
			client, err := f.ApiClient()
			if err != nil {
				return err
			}

			// Interactive mode if no name provided
			if name == "" {
				if err := huh.NewInput().
					Title("Project Name").
					Description("Enter the name for your new project").
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
					Title("Description (optional)").
					Description("Enter an optional description for the project").
					Value(&description).
					Run(); err != nil {
					return err
				}
			}

			// Create the project
			req := api.ProjectCreateRequest{
				Name:        name,
				Description: description,
			}

			project, err := client.Projects().CreateProject(f.Context(), req)
			if err != nil {
				return fmt.Errorf("failed to create project: %w", err)
			}

			// Output
			if jsonOutput {
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(project)
			}

			if ui.IsTerminal(os.Stdout) {
				fmt.Printf("✓ Created project: %s (#%d)\n", project.Name, project.ID)
			} else {
				fmt.Printf("%d\n", project.ID)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Project name")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Project description")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}
