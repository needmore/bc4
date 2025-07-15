package card

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
)

func newMoveCmd(f *factory.Factory) *cobra.Command {
	var columnName string
	var accountID string
	var projectID string

	cmd := &cobra.Command{
		Use:   "move [ID or URL]",
		Short: "Move card between columns",
		Long: `Move a card to a different column in the card table.

You can specify the card using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345")

Examples:
  bc4 card move 123 --column "In Progress"
  bc4 card move 123 --column 456
  bc4 card move https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345 --column "Done"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse card ID (could be numeric ID or URL)
			cardID, parsedURL, err := parser.ParseArgument(args[0])
			if err != nil {
				return fmt.Errorf("invalid card ID or URL: %s", args[0])
			}

			if columnName == "" {
				return fmt.Errorf("--column flag is required")
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
				if parsedURL.ResourceType != parser.ResourceTypeCard {
					return fmt.Errorf("URL is not for a card: %s", args[0])
				}
				if parsedURL.AccountID > 0 {
					f = f.WithAccount(strconv.FormatInt(parsedURL.AccountID, 10))
				}
				if parsedURL.ProjectID > 0 {
					f = f.WithProject(strconv.FormatInt(parsedURL.ProjectID, 10))
				}
			}

			// Get resolved project ID
			resolvedProjectID, err := f.ProjectID()
			if err != nil {
				return err
			}

			// Get API client from factory
			client, err := f.ApiClient()
			if err != nil {
				return err
			}
			cardOps := client.Cards()

			// First, get the card to find its current card table
			_, err = cardOps.GetCard(f.Context(), resolvedProjectID, cardID)
			if err != nil {
				return fmt.Errorf("failed to get card: %w", err)
			}

			// Get the card table to find the target column
			// We need to find the card table ID from the card's parent
			// This is a simplified implementation - in reality we'd need to traverse the parent chain
			cardTable, err := cardOps.GetProjectCardTable(f.Context(), resolvedProjectID)
			if err != nil {
				return fmt.Errorf("failed to get card table: %w", err)
			}

			// Find the target column by name or ID
			var targetColumnID int64

			// Try to parse as ID first
			if id, err := strconv.ParseInt(columnName, 10, 64); err == nil {
				targetColumnID = id
			} else {
				// Search by name
				columnNameLower := strings.ToLower(columnName)
				for _, column := range cardTable.Lists {
					if strings.ToLower(column.Title) == columnNameLower {
						targetColumnID = column.ID
						break
					}
				}
				if targetColumnID == 0 {
					return fmt.Errorf("column '%s' not found", columnName)
				}
			}

			// Move the card
			err = cardOps.MoveCard(f.Context(), resolvedProjectID, cardID, targetColumnID)
			if err != nil {
				return fmt.Errorf("failed to move card: %w", err)
			}

			// Get the column name for the success message
			var targetColumnName string
			for _, column := range cardTable.Lists {
				if column.ID == targetColumnID {
					targetColumnName = column.Title
					break
				}
			}

			fmt.Printf("✓ Moved card #%d to column '%s'\n", cardID, targetColumnName)

			return nil
		},
	}

	cmd.Flags().StringVar(&columnName, "column", "", "Target column name or ID (required)")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")
	cmd.MarkFlagRequired("column")

	return cmd
}
