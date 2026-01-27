package card

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/ui/tableprinter"
	"github.com/spf13/cobra"
)

func newTableCmd(f *factory.Factory) *cobra.Command {
	var formatJSON bool
	var accountID string
	var projectID string
	var columnFilter string
	var format string

	cmd := &cobra.Command{
		Use:   "table [ID|name]",
		Short: "View cards in a specific card table",
		Long: `View all cards in a specific card table, organized by columns.
On-hold cards are included automatically and shown with an [ON HOLD] indicator.

If no table ID or name is provided, uses the default card table if set.`,
		Args: cobra.MaximumNArgs(1),
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
			cardOps := client.Cards()

			// Get resolved IDs
			resolvedAccountID, err := f.AccountID()
			if err != nil {
				return err
			}
			resolvedProjectID, err := f.ProjectID()
			if err != nil {
				return err
			}

			// Get config for default lookups
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			// Get card table ID
			var cardTableID int64
			if len(args) > 0 {
				// Try to parse as ID first
				if id, err := strconv.ParseInt(args[0], 10, 64); err == nil {
					cardTableID = id
				} else {
					// Not implemented: search by name
					return fmt.Errorf("searching card tables by name not yet implemented")
				}
			} else {
				// Use default card table
				if acc, ok := cfg.Accounts[resolvedAccountID]; ok {
					if proj, ok := acc.ProjectDefaults[resolvedProjectID]; ok && proj.DefaultCardTable != "" {
						if id, err := strconv.ParseInt(proj.DefaultCardTable, 10, 64); err == nil {
							cardTableID = id
						}
					}
				}
				if cardTableID == 0 {
					return fmt.Errorf("no card table specified and no default card table set")
				}
			}

			// Get the card table
			cardTable, err := cardOps.GetCardTable(f.Context(), resolvedProjectID, cardTableID)
			if err != nil {
				return fmt.Errorf("failed to fetch card table: %w", err)
			}

			// Handle JSON output
			if formatJSON || format == "json" {
				// For JSON output, return all cards from all columns
				fmt.Printf("Cards in table: %s\n", cardTable.Title)
				return nil
			}

			// Create table
			table := tableprinter.New(os.Stdout)

			// Add headers
			if table.IsTTY() {
				table.AddHeader("ID", "TITLE", "COLUMN", "ASSIGNEES", "STEPS", "DUE", "UPDATED")
			} else {
				table.AddHeader("ID", "TITLE", "COLUMN", "ASSIGNEES", "STEPS", "DUE", "STATUS", "UPDATED")
			}

			totalCards := 0
			// Iterate through columns
			for _, column := range cardTable.Lists {
				// Skip if filtering by column and this doesn't match
				if columnFilter != "" && !strings.Contains(strings.ToLower(column.Title), strings.ToLower(columnFilter)) {
					continue
				}

				// Get cards in this column
				cards, err := cardOps.GetCardsInColumn(f.Context(), resolvedProjectID, column.ID)
				if err != nil {
					return fmt.Errorf("failed to fetch cards from column %s: %w", column.Title, err)
				}

				// Fetch on-hold cards if the column has them
				if column.OnHold.CardsURL != "" {
					onHoldCards, ohErr := cardOps.GetOnHoldCardsInColumn(f.Context(), column.OnHold.CardsURL)
					if ohErr == nil {
						for i := range onHoldCards {
							onHoldCards[i].IsOnHold = true
						}
						cards = append(cards, onHoldCards...)
					}
				}

				// Add each card to the table
				for _, card := range cards {
					totalCards++

					// ID
					table.AddIDField(fmt.Sprintf("%d", card.ID), card.Status)

					// Title
					title := card.Title
					if card.IsOnHold {
						title = "[ON HOLD] " + title
					}
					table.AddProjectField(title, card.Status)

					// Column with color
					columnTitle := column.Title
					if column.Color != "" && column.Color != "white" {
						// Could add color indicators here
						columnTitle = fmt.Sprintf("%s (%s)", column.Title, column.Color)
					}
					table.AddField(columnTitle)

					// Assignees
					assigneeNames := []string{}
					for _, assignee := range card.Assignees {
						assigneeNames = append(assigneeNames, assignee.Name)
					}
					table.AddField(strings.Join(assigneeNames, ", "))

					// Steps progress
					completedSteps := 0
					for _, step := range card.Steps {
						if step.Completed {
							completedSteps++
						}
					}
					if len(card.Steps) > 0 {
						table.AddField(fmt.Sprintf("%d/%d", completedSteps, len(card.Steps)))
					} else {
						table.AddField("-")
					}

					// Due date
					if card.DueOn != nil && *card.DueOn != "" {
						table.AddField(*card.DueOn)
					} else {
						table.AddField("-")
					}

					// Status for non-TTY
					if !table.IsTTY() {
						table.AddField(card.Status)
					}

					// Updated timestamp
					table.AddTimeField(card.CreatedAt, card.UpdatedAt)
					table.EndRow()
				}
			}

			// Print summary
			fmt.Printf("Showing %d cards in %s\n\n", totalCards, cardTable.Title)
			_ = table.Render()

			return nil
		},
	}

	cmd.Flags().BoolVar(&formatJSON, "json", false, "Output in JSON format")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")
	cmd.Flags().StringVar(&columnFilter, "column", "", "Filter to show only specific column")
	cmd.Flags().StringVarP(&format, "format", "f", "table", "Output format (table, json, csv)")

	return cmd
}
