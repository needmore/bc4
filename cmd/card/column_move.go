package card

import (
	"fmt"
	"strconv"

	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
)

func newColumnMoveCmd(f *factory.Factory) *cobra.Command {
	var accountID string
	var projectID string
	var cardTableID string

	cmd := &cobra.Command{
		Use:   "move [COLUMN_ID or URL]",
		Short: "Move a column to a different position",
		Long: `Move a column to a different position in the card table.

You can specify the column using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/card_tables/columns/12345")

Examples:
  bc4 card column move 123 --after 456 --table 789
  bc4 card column move 123 --before 456 --table 789
  bc4 card column move https://3.basecamp.com/1234567/buckets/89012345/card_tables/columns/12345 --after 456`,
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

			// Parse card table ID if provided
			var tableID int64
			if cardTableID != "" {
				tableID, err = strconv.ParseInt(cardTableID, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid card table ID: %s", cardTableID)
				}
			} else {
				return fmt.Errorf("--table flag is required to specify the card table ID")
			}

			// Get position flags
			after, _ := cmd.Flags().GetString("after")
			before, _ := cmd.Flags().GetString("before")

			// Validate that exactly one position flag is set
			if (after == "" && before == "") || (after != "" && before != "") {
				return fmt.Errorf("exactly one of --after or --before must be specified")
			}

			var targetID int64
			var position string
			if after != "" {
				targetID, err = strconv.ParseInt(after, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid column ID in --after: %s", after)
				}
				position = "after"
			} else {
				targetID, err = strconv.ParseInt(before, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid column ID in --before: %s", before)
				}
				position = "before"
			}

			// Move the column
			err = client.Columns().MoveColumn(f.Context(), resolvedProjectID, tableID, columnID, targetID, position)
			if err != nil {
				return fmt.Errorf("failed to move column: %w", err)
			}

			fmt.Printf("Column #%d moved\n", columnID)
			return nil
		},
	}

	cmd.Flags().String("after", "", "Move column after another column ID")
	cmd.Flags().String("before", "", "Move column before another column ID")
	cmd.Flags().StringVar(&cardTableID, "table", "", "Card table ID (required)")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")

	return cmd
}
