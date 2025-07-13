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

func newViewCmd() *cobra.Command {
	var formatJSON bool
	var accountID string
	var projectID string
	var stepsOnly bool
	var web bool

	cmd := &cobra.Command{
		Use:   "view [ID]",
		Short: "View card details including steps",
		Long:  `View detailed information about a specific card, including its description, assignees, and steps.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Parse card ID
			cardID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid card ID: %s", args[0])
			}

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

			// Handle web flag
			if web {
				// Open in browser (implementation would go here)
				fmt.Printf("Opening card %d in browser...\n", cardID)
				return nil
			}

			// Create auth client
			authClient := auth.NewClient(cfg.ClientID, cfg.ClientSecret)
			token, err := authClient.GetToken(accountID)
			if err != nil {
				return fmt.Errorf("failed to get auth token: %w", err)
			}

			// Create API client
			client := api.NewClient(accountID, token.AccessToken)

			// Get the card
			card, err := client.GetCard(ctx, projectID, cardID)
			if err != nil {
				return fmt.Errorf("failed to fetch card: %w", err)
			}

			// Handle JSON output
			if formatJSON {
				// For JSON output, return the card structure
				fmt.Printf("Card: %s\n", card.Title)
				return nil
			}

			// If steps only, show just the steps
			if stepsOnly {
				return showStepsTable(card)
			}

			// Show card header
			fmt.Printf("Card #%d: %s\n", card.ID, card.Title)
			fmt.Println(strings.Repeat("-", 50))

			// Show card details
			if card.Content != "" {
				fmt.Printf("Description: %s\n", card.Content)
			}

			// Column
			if card.Parent != nil {
				fmt.Printf("Column: %s", card.Parent.Title)
				if card.Parent.Color != "" && card.Parent.Color != "white" {
					fmt.Printf(" (%s)", card.Parent.Color)
				}
				fmt.Println()
			}

			// Assignees
			if len(card.Assignees) > 0 {
				assigneeNames := []string{}
				for _, assignee := range card.Assignees {
					assigneeNames = append(assigneeNames, assignee.Name)
				}
				fmt.Printf("Assignees: %s\n", strings.Join(assigneeNames, ", "))
			}

			// Due date
			if card.DueOn != nil && *card.DueOn != "" {
				fmt.Printf("Due: %s\n", *card.DueOn)
			}

			// Creator
			if card.Creator != nil {
				fmt.Printf("Created by: %s\n", card.Creator.Name)
			}

			// Timestamps
			fmt.Printf("Created: %s\n", card.CreatedAt.Format("2006-01-02 15:04"))
			fmt.Printf("Updated: %s\n", card.UpdatedAt.Format("2006-01-02 15:04"))

			// Show steps if any
			if len(card.Steps) > 0 {
				fmt.Printf("\nSteps (%d):\n", len(card.Steps))
				fmt.Println(strings.Repeat("-", 50))
				return showStepsTable(card)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&formatJSON, "json", false, "Output in JSON format")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")
	cmd.Flags().BoolVar(&stepsOnly, "steps-only", false, "Show only the steps list")
	cmd.Flags().BoolVarP(&web, "web", "w", false, "Open card in web browser")

	return cmd
}

func showStepsTable(card *api.Card) error {
	table := tableprinter.New(os.Stdout)

	// Add headers
	if table.IsTTY() {
		table.AddHeader("", "ID", "TITLE", "ASSIGNEES", "DUE", "UPDATED")
	} else {
		table.AddHeader("STATUS", "ID", "TITLE", "ASSIGNEES", "DUE", "UPDATED")
	}

	// Add each step
	for _, step := range card.Steps {
		// Status indicator
		if table.IsTTY() {
			table.AddStatusField(step.Completed)
		} else {
			if step.Completed {
				table.AddField("completed")
			} else {
				table.AddField("incomplete")
			}
		}

		// Step ID
		stepID := fmt.Sprintf("%d", step.ID)
		table.AddIDField(stepID, step.Status)

		// Title
		table.AddTodoField(step.Title, step.Completed)

		// Assignees
		assigneeNames := []string{}
		for _, assignee := range step.Assignees {
			assigneeNames = append(assigneeNames, assignee.Name)
		}
		table.AddField(strings.Join(assigneeNames, ", "))

		// Due date
		if step.DueOn != nil && *step.DueOn != "" {
			table.AddField(*step.DueOn)
		} else {
			table.AddField("-")
		}

		// Updated timestamp
		table.AddTimeField(step.CreatedAt, step.UpdatedAt)
		table.EndRow()
	}

	table.Render()
	return nil
}
