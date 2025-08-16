package card

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/needmore/bc4/internal/ui/tableprinter"
	"github.com/spf13/cobra"
)

func newColumnListCmd(f *factory.Factory) *cobra.Command {
	var accountID string
	var projectID string

	cmd := &cobra.Command{
		Use:   "list [CARD_TABLE_ID or URL]",
		Short: "List all columns in a card table",
		Long: `List all columns in the specified card table.

You can specify the card table using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/card_tables/12345")

Examples:
  bc4 card column list 123
  bc4 card column list 123 --format json
  bc4 card column list https://3.basecamp.com/1234567/buckets/89012345/card_tables/12345`,
		Args: cobra.ExactArgs(1),
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

			// Get the card table with its columns
			cardTable, err := client.Cards().GetCardTable(f.Context(), resolvedProjectID, cardTableID)
			if err != nil {
				return fmt.Errorf("failed to get card table: %w", err)
			}

			// Handle different output formats
			format, _ := cmd.Flags().GetString("format")
			switch format {
			case "json":
				output, err := json.MarshalIndent(cardTable.Lists, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to format JSON: %w", err)
				}
				fmt.Println(string(output))

			case "csv":
				// Output comma-separated values using proper CSV writer
				writer := csv.NewWriter(os.Stdout)
				defer writer.Flush()

				// Write header
				if err := writer.Write([]string{"ID", "TITLE", "DESCRIPTION", "CARDS_COUNT"}); err != nil {
					return fmt.Errorf("failed to write CSV header: %w", err)
				}

				// Write data rows
				for _, column := range cardTable.Lists {
					record := []string{
						strconv.FormatInt(column.ID, 10),
						column.Title,
						"", // Description
						strconv.Itoa(column.CardsCount),
					}
					if err := writer.Write(record); err != nil {
						return fmt.Errorf("failed to write CSV record: %w", err)
					}
				}

			default: // table format
				table := tableprinter.New(os.Stdout)

				// Add headers
				if table.IsTTY() {
					table.AddHeader("ID", "TITLE", "DESCRIPTION", "CARDS")
				} else {
					table.AddHeader("ID", "TITLE", "DESCRIPTION", "CARDS_COUNT")
				}

				// Add rows
				for _, column := range cardTable.Lists {
					table.AddIDField(fmt.Sprintf("%d", column.ID), column.Status)
					table.AddField(column.Title)
					table.AddField("")
					table.AddField(fmt.Sprintf("%d", column.CardsCount))
					table.EndRow()
				}

				_ = table.Render()
			}

			return nil
		},
	}

	cmd.Flags().String("format", "table", "Output format: table, json, csv")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")

	return cmd
}
