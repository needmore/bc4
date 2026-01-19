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

func newUnarchiveCmd(f *factory.Factory) *cobra.Command {
	var accountID string
	var skipConfirm bool

	cmd := &cobra.Command{
		Use:   "unarchive <project-id|url>",
		Short: "Restore an archived project",
		Long: `Restore an archived project back to active status.

You can specify the project using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/projects/89012345")

Examples:
  bc4 project unarchive 12345
  bc4 project unarchive 12345 --yes           # Skip confirmation
  bc4 project unarchive https://3.basecamp.com/1234567/projects/89012345`,
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

			// Fetch project first to show what will be unarchived
			project, err := client.Projects().GetProject(f.Context(), projectID)
			if err != nil {
				return fmt.Errorf("failed to fetch project: %w", err)
			}

			// Confirmation prompt unless skipped
			if !skipConfirm {
				var confirm bool
				if err := huh.NewConfirm().
					Title(fmt.Sprintf("Restore project \"%s\"?", project.Name)).
					Description("The project will be restored to active projects.").
					Affirmative("Restore").
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

			// Unarchive the project
			if err := client.Projects().UnarchiveProject(f.Context(), projectID); err != nil {
				return fmt.Errorf("failed to unarchive project: %w", err)
			}

			// Output
			if ui.IsTerminal(os.Stdout) {
				fmt.Printf("✓ Restored project: %s (#%d)\n", project.Name, project.ID)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
