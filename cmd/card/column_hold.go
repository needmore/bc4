package card

import (
	"fmt"
	"strconv"

	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
)

func newColumnHoldCmd(f *factory.Factory) *cobra.Command {
	var accountID string
	var projectID string

	cmd := &cobra.Command{
		Use:     "hold [COLUMN_ID or URL]",
		Aliases: []string{"on-hold"},
		Short:   "Mark a column as on-hold",
		Long: `Mark a column as on-hold to indicate paused or blocked work.

On-hold columns are visually distinguished in Basecamp's kanban board view.

You can specify the column using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/card_tables/columns/12345")

Examples:
  bc4 card column hold 123
  bc4 card column on-hold 123
  bc4 card column hold https://3.basecamp.com/.../columns/12345`,
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

			// Set the column on-hold
			err = client.Columns().SetColumnOnHold(f.Context(), resolvedProjectID, columnID)
			if err != nil {
				return fmt.Errorf("failed to set column on-hold: %w", err)
			}

			fmt.Printf("Column #%d marked as on-hold\n", columnID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")

	return cmd
}
