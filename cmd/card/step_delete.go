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

// newStepDeleteCmd creates the step delete command
func newStepDeleteCmd() *cobra.Command {
	var accountID string
	var projectID string

	cmd := &cobra.Command{
		Use:   "delete [CARD_ID or URL] STEP_ID",
		Short: "Delete a step from a card",
		Long: `Delete a step (subtask) from a card.

You can specify the card using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345")

Examples:
  bc4 card step delete 123 456
  bc4 card step delete 123 456 --force
  bc4 card step delete https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345 456`,
		Args: cobra.ExactArgs(2),
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

			// Parse step ID
			stepID, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid step ID: %s", args[1])
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

			// TODO: Implement step delete functionality
			fmt.Printf("Would delete step ID %d from card ID %d\n", stepID, cardID)
			return fmt.Errorf("step delete not yet implemented")
		},
	}

	// TODO: Add flags for force delete
	cmd.Flags().Bool("force", false, "Delete without confirmation")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")

	return cmd
}
