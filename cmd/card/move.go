package card

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
)

func newMoveCmd(f *factory.Factory) *cobra.Command {
	var columnName string
	var accountID string
	var projectID string
	var onHold bool

	cmd := &cobra.Command{
		Use:   "move [ID or URL]",
		Short: "Move card between columns",
		Long: `Move a card to a different column in the card table.

You can specify the card using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345")

Use --on-hold to move a card to the on-hold section of its current column
(or target column if --column is also specified).

Examples:
  bc4 card move 123 --column "In Progress"
  bc4 card move 123 --column 456
  bc4 card move 123 --on-hold
  bc4 card move 123 --column "Developing" --on-hold
  bc4 card move https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345 --column "Done"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse card ID (could be numeric ID or URL)
			cardID, parsedURL, err := parser.ParseArgument(args[0])
			if err != nil {
				return fmt.Errorf("invalid card ID or URL: %s", args[0])
			}

			if columnName == "" && !onHold {
				return fmt.Errorf("--column flag is required (or use --on-hold)")
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

			// First, get the card to find its current location
			card, err := cardOps.GetCard(f.Context(), resolvedProjectID, cardID)
			if err != nil {
				return fmt.Errorf("failed to get card: %w", err)
			}

			// Get all card tables in the project to find the one containing this card
			cardTables, err := cardOps.GetAllProjectCardTables(f.Context(), resolvedProjectID)
			if err != nil {
				return fmt.Errorf("failed to get card tables: %w", err)
			}

			// Find which card table contains the card's current column
			var currentCardTable *api.CardTable
			if card.Parent != nil {
				for _, table := range cardTables {
					for _, column := range table.Lists {
						if column.ID == card.Parent.ID {
							currentCardTable = table
							break
						}
					}
					if currentCardTable != nil {
						break
					}
				}
			}

			// If we couldn't find the card table from the parent, use the default one
			if currentCardTable == nil {
				if len(cardTables) > 0 {
					currentCardTable = cardTables[0]
				} else {
					return fmt.Errorf("no card tables found in project")
				}
			}

			// Handle --on-hold: move to the on-hold section
			if onHold {
				targetColumn, err := findColumn(currentCardTable, columnName, card)
				if err != nil {
					return err
				}
				if targetColumn.OnHold.ID == 0 {
					return fmt.Errorf("column '%s' does not have an on-hold section", targetColumn.Title)
				}
				if err := cardOps.MoveCard(f.Context(), resolvedProjectID, cardID, targetColumn.OnHold.ID); err != nil {
					return fmt.Errorf("failed to move card to on-hold: %w", err)
				}
				fmt.Printf("✓ Moved card #%d to on-hold in column '%s'\n", cardID, targetColumn.Title)
				return nil
			}

			// Find the target column by name or ID within the same card table
			var targetColumnID int64

			// Try to parse as ID first
			if id, err := strconv.ParseInt(columnName, 10, 64); err == nil {
				// Validate that the column ID exists in the current card table
				found := false
				for _, column := range currentCardTable.Lists {
					if column.ID == id {
						targetColumnID = id
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("column ID %d not found in card table '%s'", id, currentCardTable.Title)
				}
			} else {
				// Search by name in the same card table
				columnNameLower := strings.ToLower(columnName)
				for _, column := range currentCardTable.Lists {
					if strings.ToLower(column.Title) == columnNameLower {
						targetColumnID = column.ID
						break
					}
				}
				if targetColumnID == 0 {
					return fmt.Errorf("column '%s' not found in card table '%s'", columnName, currentCardTable.Title)
				}
			}

			// Move the card
			err = cardOps.MoveCard(f.Context(), resolvedProjectID, cardID, targetColumnID)
			if err != nil {
				return fmt.Errorf("failed to move card: %w", err)
			}

			// Get the column name for the success message
			var targetColumnName string
			for _, column := range currentCardTable.Lists {
				if column.ID == targetColumnID {
					targetColumnName = column.Title
					break
				}
			}

			fmt.Printf("✓ Moved card #%d to column '%s' on card table '%s'\n", cardID, targetColumnName, currentCardTable.Title)

			return nil
		},
	}

	cmd.Flags().StringVar(&columnName, "column", "", "Target column name or ID")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")
	cmd.Flags().BoolVar(&onHold, "on-hold", false, "Move card to the on-hold section of its current (or target) column")

	return cmd
}

// findColumn resolves the target column from --column flag or falls back to the card's current column.
func findColumn(cardTable *api.CardTable, columnName string, card *api.Card) (*api.Column, error) {
	if columnName != "" {
		// Try as ID first
		if id, err := strconv.ParseInt(columnName, 10, 64); err == nil {
			for i := range cardTable.Lists {
				if cardTable.Lists[i].ID == id {
					return &cardTable.Lists[i], nil
				}
			}
			return nil, fmt.Errorf("column ID %d not found in card table '%s'", id, cardTable.Title)
		}
		// Search by name
		columnNameLower := strings.ToLower(columnName)
		for i := range cardTable.Lists {
			if strings.ToLower(cardTable.Lists[i].Title) == columnNameLower {
				return &cardTable.Lists[i], nil
			}
		}
		return nil, fmt.Errorf("column '%s' not found in card table '%s'", columnName, cardTable.Title)
	}

	// No --column specified, use card's current column
	if card.Parent == nil {
		return nil, fmt.Errorf("card has no parent column")
	}
	for i := range cardTable.Lists {
		if cardTable.Lists[i].ID == card.Parent.ID {
			return &cardTable.Lists[i], nil
		}
	}
	return nil, fmt.Errorf("could not find card's current column in card table")
}
