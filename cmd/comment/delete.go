package comment

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
	var projectID string
	var skipConfirm bool

	cmd := &cobra.Command{
		Use:   "delete <comment-id|url>",
		Short: "Delete a comment",
		Long:  `Delete a comment. This operation cannot be undone.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Apply overrides if specified
			if accountID != "" {
				f = f.WithAccount(accountID)
			}
			if projectID != "" {
				f = f.WithProject(projectID)
			}

			// Get API client from factory
			client, err := f.ApiClient()
			if err != nil {
				return err
			}

			var commentID int64
			var projectID string

			// Parse the argument - could be a URL or ID
			if parser.IsBasecampURL(args[0]) {
				parsed, err := parser.ParseBasecampURL(args[0])
				if err != nil {
					return fmt.Errorf("invalid Basecamp URL: %w", err)
				}
				if parsed.ResourceType != parser.ResourceTypeComment {
					return fmt.Errorf("URL is not a comment URL: %s", args[0])
				}
				commentID = parsed.ResourceID
				projectID = strconv.FormatInt(parsed.ProjectID, 10)
			} else {
				// It's just an ID, we need the project ID from config
				commentID, err = strconv.ParseInt(args[0], 10, 64)
				if err != nil {
					return fmt.Errorf("invalid comment ID: %s", args[0])
				}
				projectID, err = f.ProjectID()
				if err != nil {
					return err
				}
			}

			// Get the comment first to show what will be deleted
			comment, err := client.GetComment(f.Context(), projectID, commentID)
			if err != nil {
				return err
			}

			// Confirmation prompt unless skipped
			if !skipConfirm {
				var confirm bool
				if err := huh.NewConfirm().
					Title(fmt.Sprintf("Delete comment #%d?", commentID)).
					Description(fmt.Sprintf("By %s on %s", comment.Creator.Name, comment.CreatedAt.Format("Jan 2, 2006"))).
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

			// Trash the comment via Basecamp recordings API
			path := fmt.Sprintf("/buckets/%s/recordings/%d/status/trashed.json", projectID, commentID)
			if err := client.Put(path, nil, nil); err != nil {
				return fmt.Errorf("failed to delete comment: %w", err)
			}

			// Output
			if ui.IsTerminal(os.Stdout) {
				fmt.Printf("✓ Deleted comment #%d\n", commentID)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")
	cmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
