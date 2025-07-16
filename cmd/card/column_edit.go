package card

import (
	"fmt"
	"strconv"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
)

func newColumnEditCmd(f *factory.Factory) *cobra.Command {
	var accountID string
	var projectID string

	cmd := &cobra.Command{
		Use:   "edit [COLUMN_ID or URL]",
		Short: "Edit the name and description of a column",
		Long: `Edit the name and description of a column.

You can specify the column using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/card_tables/columns/12345")

Examples:
  bc4 card column edit 123 --title "Done"
  bc4 card column edit 123 --description "Completed tasks"
  bc4 card column edit 123 --title "In Review" --description "Items awaiting review"
  bc4 card column edit https://3.basecamp.com/1234567/buckets/89012345/card_tables/columns/12345 --title "Done"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse column ID (could be numeric ID or URL)
			columnID, parsedURL, err := parser.ParseArgument(args[0])
			if err != nil {
				return fmt.Errorf("invalid column ID or URL: %s", args[0])
			}

			// Apply overrides if specified
			if accountID != "" {
				f = f.WithAccount(accountID)
			}
			if projectID != "" {
				f = f.WithProject(projectID)
			}

			// If a URL was parsed, override account and project IDs if provided
			if parsedURL != nil {
				if parsedURL.ResourceType != parser.ResourceTypeColumn {
					return fmt.Errorf("URL is not for a column: %s", args[0])
				}
				if parsedURL.AccountID > 0 {
					f = f.WithAccount(strconv.FormatInt(parsedURL.AccountID, 10))
				}
				if parsedURL.ProjectID > 0 {
					f = f.WithProject(strconv.FormatInt(parsedURL.ProjectID, 10))
				}
			}

			// Get API client from factory
			client, err := f.ApiClient()
			if err != nil {
				return err
			}

			// Get resolved project ID
			resolvedProjectID, err := f.ProjectID()
			if err != nil {
				return err
			}

			// Get title and description from flags
			title, _ := cmd.Flags().GetString("title")
			description, _ := cmd.Flags().GetString("description")

			// Validate that at least one field is being updated
			if title == "" && description == "" {
				return fmt.Errorf("at least one of --title or --description must be specified")
			}

			// Create the update request
			req := api.ColumnUpdateRequest{
				Title:       title,
				Description: description,
			}

			// Update the column
			column, err := client.Columns().UpdateColumn(f.Context(), resolvedProjectID, columnID, req)
			if err != nil {
				return fmt.Errorf("failed to update column: %w", err)
			}

			// Output the column ID
			fmt.Printf("Column #%d updated\n", column.ID)
			return nil
		},
	}

	cmd.Flags().String("title", "", "New title for the column")
	cmd.Flags().String("description", "", "New description for the column")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")

	return cmd
}
