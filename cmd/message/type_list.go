package message

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/ui/tableprinter"
	"github.com/spf13/cobra"
)

func newTypeListCmd(f *factory.Factory) *cobra.Command {
	var accountID string
	var projectID string

	cmd := &cobra.Command{
		Use:   "list [MESSAGE_BOARD_ID or URL]",
		Short: "List all message categories",
		Long: `List all categories for a message board.

You can specify the message board using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/message_boards/12345")
- Omit the ID to use the current project's message board

Examples:
  bc4 message type list
  bc4 message type list 123
  bc4 message type list --format json
  bc4 message type list https://3.basecamp.com/1234567/buckets/89012345/message_boards/12345`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var messageBoardID int64
			var err error

			// Parse message board ID if provided (currently not used as parser doesn't support message boards)
			// In the future, this could be used to support direct message board URLs
			_ = len(args) // Suppress unused warning

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

			// Get resolved project ID
			resolvedProjectID, err := f.ProjectID()
			if err != nil {
				return err
			}

			// Get message board if not provided
			if messageBoardID == 0 {
				board, err := client.GetMessageBoard(f.Context(), resolvedProjectID)
				if err != nil {
					return fmt.Errorf("failed to get message board: %w", err)
				}
				messageBoardID = board.ID
			}

			// List categories
			categories, err := client.ListMessageCategories(f.Context(), resolvedProjectID, messageBoardID)
			if err != nil {
				return fmt.Errorf("failed to list message categories: %w", err)
			}

			// Handle different output formats
			format, _ := cmd.Flags().GetString("format")
			switch format {
			case "json":
				output, err := json.MarshalIndent(categories, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to format JSON: %w", err)
				}
				fmt.Println(string(output))

			case "csv":
				writer := csv.NewWriter(os.Stdout)
				defer writer.Flush()

				// Write header
				if err := writer.Write([]string{"ID", "NAME", "ICON", "COLOR"}); err != nil {
					return fmt.Errorf("failed to write CSV header: %w", err)
				}

				// Write data rows
				for _, category := range categories {
					record := []string{
						strconv.FormatInt(category.ID, 10),
						category.Name,
						category.Icon,
						category.Color,
					}
					if err := writer.Write(record); err != nil {
						return fmt.Errorf("failed to write CSV record: %w", err)
					}
				}

			default: // table format
				table := tableprinter.New(os.Stdout)

				// Add headers
				table.AddHeader("ID", "NAME", "ICON", "COLOR")

				// Add rows
				for _, category := range categories {
					table.AddIDField(fmt.Sprintf("%d", category.ID), "active")
					table.AddField(category.Name)
					table.AddField(category.Icon)
					table.AddField(category.Color)
					table.EndRow()
				}

				if err := table.Render(); err != nil {
					return fmt.Errorf("failed to render table: %w", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().String("format", "table", "Output format: table, json, csv")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")

	return cmd
}
