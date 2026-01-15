package comment

import (
	"fmt"
	"os"
	"strconv"

	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/needmore/bc4/internal/ui/tableprinter"
	"github.com/spf13/cobra"
)

func newListCmd(f *factory.Factory) *cobra.Command {
	var accountID string
	var projectID string

	cmd := &cobra.Command{
		Use:     "list <recording-id|url>",
		Short:   "List comments on a recording",
		Long:    `List all comments on a Basecamp recording (todo, message, document, or card).`,
		Aliases: []string{"ls"},
		Args:    cobra.ExactArgs(1),
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

			var recordingID int64
			var projectID string

			// Parse the argument - could be a URL or ID
			if parser.IsBasecampURL(args[0]) {
				parsed, err := parser.ParseBasecampURL(args[0])
				if err != nil {
					return fmt.Errorf("invalid Basecamp URL: %w", err)
				}
				recordingID = parsed.ResourceID
				projectID = strconv.FormatInt(parsed.ProjectID, 10)
			} else {
				// It's just an ID, we need the project ID from config
				recordingID, err = strconv.ParseInt(args[0], 10, 64)
				if err != nil {
					return fmt.Errorf("invalid recording ID: %s", args[0])
				}
				projectID, err = f.ProjectID()
				if err != nil {
					return err
				}
			}

			// Get comments
			comments, err := client.ListComments(f.Context(), projectID, recordingID)
			if err != nil {
				return err
			}

			if len(comments) == 0 {
				fmt.Println("No comments found")
				return nil
			}

			// Create table
			table := tableprinter.New(os.Stdout)
			table.AddHeader("ID", "AUTHOR", "CREATED", "PREVIEW")

			for _, comment := range comments {
				// Create a short preview
				preview := comment.Content
				if len(preview) > 60 {
					preview = preview[:60] + "..."
				}

				table.AddIDField(strconv.FormatInt(comment.ID, 10), comment.Status)
				table.AddField(comment.Creator.Name)
				table.AddField(comment.CreatedAt.Format("Jan 2, 2006"))
				table.AddField(preview)
				table.EndRow()
			}

			return table.Render()
		},
	}

	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")

	return cmd
}
