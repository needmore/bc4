package card

import (
	"context"
	"fmt"
	"strconv"

	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
)

// newStepListCmd creates the step list command
func newStepListCmd() *cobra.Command {
	var accountID string
	var projectID string

	cmd := &cobra.Command{
		Use:   "list [CARD_ID or URL]",
		Short: "List all steps in a card",
		Long: `List all steps (subtasks) in a card, showing their status and assignees.

You can specify the card using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345")

Examples:
  bc4 card step list 123
  bc4 card step list 123 --completed
  bc4 card step list 123 --format json
  bc4 card step list https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = context.Background()

			// Load config
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			// Parse card ID (could be numeric ID or URL)
			cardID, parsedURL, err := parser.ParseArgument(args[0])
			if err != nil {
				return fmt.Errorf("invalid card ID or URL: %s", args[0])
			}

			// Check authentication
			if cfg.DefaultAccount == "" {
				return fmt.Errorf("not authenticated. Run 'bc4' to set up authentication")
			}

			// Get account ID
			if accountID == "" {
				accountID = cfg.DefaultAccount
			}

			// Get project ID
			if projectID == "" {
				projectID = cfg.DefaultProject
				if projectID == "" {
					// Check for account-specific default project
					if acc, ok := cfg.Accounts[accountID]; ok && acc.DefaultProject != "" {
						projectID = acc.DefaultProject
					}
				}
			}

			// If a URL was parsed, override account and project IDs if provided
			if parsedURL != nil {
				if parsedURL.ResourceType != parser.ResourceTypeCard {
					return fmt.Errorf("URL is not for a card: %s", args[0])
				}
				if parsedURL.AccountID > 0 {
					accountID = strconv.FormatInt(parsedURL.AccountID, 10)
				}
				if parsedURL.ProjectID > 0 {
					projectID = strconv.FormatInt(parsedURL.ProjectID, 10)
				}
			}

			// Validate we have required IDs
			if accountID == "" {
				return fmt.Errorf("no account specified and no default account set")
			}
			if projectID == "" {
				return fmt.Errorf("no project specified and no default project set")
			}

			// Create auth client
			authClient := auth.NewClient(cfg.ClientID, cfg.ClientSecret)
			_, err = authClient.GetToken(accountID)
			if err != nil {
				return fmt.Errorf("failed to get auth token: %w", err)
			}

			// TODO: Implement step list functionality
			fmt.Printf("Would list steps for card ID %d\n", cardID)
			return fmt.Errorf("step list not yet implemented")
		},
	}

	// TODO: Add flags for filtering and formatting
	cmd.Flags().Bool("completed", false, "Show only completed steps")
	cmd.Flags().Bool("pending", false, "Show only pending steps")
	cmd.Flags().String("assignee", "", "Filter by assignee")
	cmd.Flags().String("format", "table", "Output format: table, json, csv")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")

	return cmd
}
