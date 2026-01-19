package project

import (
	"fmt"
	"os"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/needmore/bc4/internal/ui"
	"github.com/spf13/cobra"
)

func newDeleteCmd(f *factory.Factory) *cobra.Command {
	var accountID string
	var skipConfirm bool

	cmd := &cobra.Command{
		Use:   "delete <project-id|url>",
		Short: "Delete/trash a project",
		Long: `Delete a project by moving it to the trash.

This action can be undone by an admin within 25 days from the Basecamp web interface.

You can specify the project using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/projects/89012345")

Examples:
  bc4 project delete 12345
  bc4 project delete 12345 --yes              # Skip confirmation
  bc4 project delete https://3.basecamp.com/1234567/projects/89012345`,
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

			// Fetch project first to show what will be deleted
			project, err := client.Projects().GetProject(f.Context(), projectID)
			if err != nil {
				return fmt.Errorf("failed to fetch project: %w", err)
			}

			// Confirmation prompt unless skipped
			if !skipConfirm {
				var confirm bool
				if err := huh.NewConfirm().
					Title(fmt.Sprintf("Delete project \"%s\"?", project.Name)).
					Description("This will move the project to trash. It can be restored within 25 days by an admin.").
					Affirmative("Delete").
					Negative("Cancel").
					Value(&confirm).
					Run(); err != nil {
					return err
				}

				if !confirm {
					fmt.Println("Canceled")
					return nil
				}
			}

			// Delete the project
			if err := client.Projects().DeleteProject(f.Context(), projectID); err != nil {
				return fmt.Errorf("failed to delete project: %w", err)
			}

			// Output
			if ui.IsTerminal(os.Stdout) {
				fmt.Printf("✓ Deleted project: %s (#%d)\n", project.Name, project.ID)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
