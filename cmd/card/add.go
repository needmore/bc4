package card

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
	"github.com/spf13/cobra"
)

func newAddCmd() *cobra.Command {
	var accountID string
	var projectID string
	var tableID string
	var columnName string
	var assignees []string
	var steps []string
	var dueOn string
	var description string

	cmd := &cobra.Command{
		Use:   "add \"Title\"",
		Short: "Quick card creation",
		Long: `Create a new card in the default card table's first column.
		
Use flags to specify table, column, assignees, and initial steps.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			title := args[0]

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
			client := api.NewClient(accountID, token.AccessToken)

			// Get card table ID
			var cardTableID int64
			if tableID != "" {
				// Parse specified table ID
				if id, err := strconv.ParseInt(tableID, 10, 64); err == nil {
					cardTableID = id
				} else {
					// Search by name not implemented yet
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
					// No default set, get the project's card table
					cardTable, err := client.GetProjectCardTable(ctx, projectID)
					if err != nil {
						return fmt.Errorf("failed to fetch card table: %w", err)
					}
					cardTableID = cardTable.ID
				}
			}

			// Get the card table to find columns
			cardTable, err := client.GetCardTable(ctx, projectID, cardTableID)
			if err != nil {
				return fmt.Errorf("failed to fetch card table: %w", err)
			}

			// Find the target column
			var targetColumn *api.Column
			if columnName != "" {
				// Find column by name
				for i := range cardTable.Lists {
					if strings.Contains(strings.ToLower(cardTable.Lists[i].Title), strings.ToLower(columnName)) {
						targetColumn = &cardTable.Lists[i]
						break
					}
				}
				if targetColumn == nil {
					return fmt.Errorf("column '%s' not found", columnName)
				}
			} else {
				// Use first non-triage column (usually the second column)
				for i := range cardTable.Lists {
					if cardTable.Lists[i].Type != "Kanban::Triage" {
						targetColumn = &cardTable.Lists[i]
						break
					}
				}
				if targetColumn == nil && len(cardTable.Lists) > 0 {
					// Fallback to first column
					targetColumn = &cardTable.Lists[0]
				}
			}

			if targetColumn == nil {
				return fmt.Errorf("no suitable column found in card table")
			}

			// Create the card
			req := api.CardCreateRequest{
				Title:   title,
				Content: description,
			}
			if dueOn != "" {
				req.DueOn = &dueOn
			}

			card, err := client.CreateCard(ctx, projectID, targetColumn.ID, req)
			if err != nil {
				return fmt.Errorf("failed to create card: %w", err)
			}

			fmt.Printf("Created card #%d: %s in column '%s'\n", card.ID, card.Title, targetColumn.Title)

			// Handle assignees - parse as IDs and update the card
			if len(assignees) > 0 {
				var assigneeIDs []int64
				for _, assignee := range assignees {
					// Try to parse as ID
					if id, err := strconv.ParseInt(assignee, 10, 64); err == nil {
						assigneeIDs = append(assigneeIDs, id)
					} else {
						fmt.Printf("Warning: '%s' is not a valid user ID, skipping\n", assignee)
					}
				}
				if len(assigneeIDs) > 0 {
					updateReq := api.CardUpdateRequest{
						AssigneeIDs: assigneeIDs,
					}
					_, err := client.UpdateCard(ctx, projectID, card.ID, updateReq)
					if err != nil {
						fmt.Printf("Warning: failed to assign users: %v\n", err)
					} else {
						fmt.Printf("Assigned %d user(s) to the card\n", len(assigneeIDs))
					}
				}
			}

			// Add steps if provided
			if len(steps) > 0 {
				fmt.Printf("Adding %d steps...\n", len(steps))
				for _, stepTitle := range steps {
					stepReq := api.StepCreateRequest{
						Title: stepTitle,
					}
					_, err := client.CreateStep(ctx, projectID, card.ID, stepReq)
					if err != nil {
						fmt.Printf("Warning: failed to add step '%s': %v\n", stepTitle, err)
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")
	cmd.Flags().StringVar(&tableID, "table", "", "Specify card table ID")
	cmd.Flags().StringVar(&columnName, "column", "", "Target column name")
	cmd.Flags().StringSliceVar(&assignees, "assign", []string{}, "Add assignees by user ID (comma-separated)")
	cmd.Flags().StringSliceVar(&steps, "step", []string{}, "Add steps (can be used multiple times)")
	cmd.Flags().StringVar(&dueOn, "due", "", "Set due date (YYYY-MM-DD)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Card description")

	return cmd
}
