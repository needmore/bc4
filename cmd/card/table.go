package card

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
	"github.com/needmore/bc4/internal/ui/tableprinter"
	"github.com/spf13/cobra"
)

func newTableCmd() *cobra.Command {
	var formatJSON bool
	var accountID string
	var projectID string
	var columnFilter string
	var format string

	cmd := &cobra.Command{
		Use:   "table [ID|name]",
		Short: "View cards in a specific card table",
		Long: `View all cards in a specific card table, organized by columns.
		
If no table ID or name is provided, uses the default card table if set.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Load config
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			// Check authentication
			if cfg.DefaultAccount == "" {
				return fmt.Errorf("not authenticated. Run 'bc4' to set up authentication")
			}

			// Get account ID
			if accountID == "" {
				accountID = cfg.DefaultAccount
			}
			if accountID == "" {
				return fmt.Errorf("no account specified and no default account set")
			}

			// Get project ID
			if projectID == "" {
				projectID = cfg.DefaultProject
			}
			if projectID == "" {
				// Check for account-specific default project
				if acc, ok := cfg.Accounts[accountID]; ok && acc.DefaultProject != "" {
					projectID = acc.DefaultProject
				}
			}
			if projectID == "" {
				return fmt.Errorf("no project specified and no default project set")
			}

			// Create auth client
			authClient := auth.NewClient(cfg.ClientID, cfg.ClientSecret)
			token, err := authClient.GetToken(accountID)
			if err != nil {
				return fmt.Errorf("failed to get auth token: %w", err)
			}

			// Create API client
			client := api.NewModularClient(accountID, token.AccessToken)
			cardOps := client.Cards()

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
				if acc, ok := cfg.Accounts[accountID]; ok {
					if proj, ok := acc.ProjectDefaults[projectID]; ok && proj.DefaultCardTable != "" {
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
			cardTable, err := cardOps.GetCardTable(ctx, projectID, cardTableID)
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
				cards, err := cardOps.GetCardsInColumn(ctx, projectID, column.ID)
				if err != nil {
					return fmt.Errorf("failed to fetch cards from column %s: %w", column.Title, err)
				}

				// Add each card to the table
				for _, card := range cards {
					totalCards++

					// ID
					table.AddIDField(fmt.Sprintf("%d", card.ID), card.Status)

					// Title
					table.AddProjectField(card.Title, card.Status)

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
			table.Render()

			return nil
		},
	}

	cmd.Flags().BoolVar(&formatJSON, "json", false, "Output in JSON format")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")
	cmd.Flags().StringVar(&columnFilter, "column", "", "Filter to show only specific column")
	cmd.Flags().StringVarP(&format, "format", "f", "table", "Output format (table, json, tsv)")

	return cmd
}
