package card

import (
	"fmt"
	"strconv"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
)

func newColumnCreateCmd(f *factory.Factory) *cobra.Command {
	var accountID string
	var projectID string

	cmd := &cobra.Command{
		Use:   "create [CARD_TABLE_ID or URL] TITLE",
		Short: "Create a new column in a card table",
		Long: `Create a new column in a card table.

You can specify the card table using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/card_tables/12345")

Examples:
  bc4 card column create 123 "In Progress"
  bc4 card column create 123 "Done" --description "Completed items"
  bc4 card column create https://3.basecamp.com/1234567/buckets/89012345/card_tables/12345 "Review"`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse card table ID (could be numeric ID or URL)
			cardTableID, parsedURL, err := parser.ParseArgument(args[0])
			if err != nil {
				return fmt.Errorf("invalid card table ID or URL: %s", args[0])
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
				if parsedURL.ResourceType != parser.ResourceTypeCardTable {
					return fmt.Errorf("URL is not for a card table: %s", args[0])
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

			// Get description from flag
			description, _ := cmd.Flags().GetString("description")

			// Create the column request
			req := api.ColumnCreateRequest{
				Title:       args[1],
				Description: description,
			}

			// Create the column
			column, err := client.Columns().CreateColumn(f.Context(), resolvedProjectID, cardTableID, req)
			if err != nil {
				return fmt.Errorf("failed to create column: %w", err)
			}

			// Output the column ID
			fmt.Printf("#%d\n", column.ID)
			return nil
		},
	}

	cmd.Flags().String("description", "", "Description for the column")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")

	return cmd
}
