package card

import (
	"context"
	"fmt"
	"strconv"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
)

// newStepCheckCmd creates the step check command
func newStepCheckCmd() *cobra.Command {
	var accountID string
	var projectID string
	var note string

	cmd := &cobra.Command{
		Use:   "check [CARD_ID or URL] [STEP_ID or URL]",
		Short: "Mark a step as completed",
		Long: `Mark a step (subtask) as completed.

You can specify the card and step using either:
- Numeric IDs (e.g., "123 456" for card 123, step 456)
- A Basecamp step URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345/steps/67890")

Examples:
  bc4 card step check 123 456
  bc4 card step check 123 456 --note "Fixed in PR #789"
  bc4 card step check https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345/steps/67890`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			var cardID, stepID int64
			var parsedURL *parser.ParsedURL

			// Parse arguments - could be card ID + step ID, or a single step URL
			if len(args) == 1 {
				// Single argument - must be a step URL
				var err error
				stepID, parsedURL, err = parser.ParseArgument(args[0])
				if err != nil {
					return fmt.Errorf("invalid step URL: %s", args[0])
				}
				if parsedURL == nil {
					return fmt.Errorf("when providing a single argument, it must be a Basecamp step URL")
				}
				if parsedURL.ResourceType != parser.ResourceTypeStep {
					return fmt.Errorf("URL is not for a step: %s", args[0])
				}
				cardID = parsedURL.ParentID
			} else {
				// Two arguments - card ID and step ID
				var err error
				cardID, _, err = parser.ParseArgument(args[0])
				if err != nil {
					return fmt.Errorf("invalid card ID or URL: %s", args[0])
				}

				stepID, parsedURL, err = parser.ParseArgument(args[1])
				if err != nil {
					return fmt.Errorf("invalid step ID or URL: %s", args[1])
				}

				// If step was provided as URL, validate and extract IDs
				if parsedURL != nil {
					if parsedURL.ResourceType != parser.ResourceTypeStep {
						return fmt.Errorf("URL is not for a step: %s", args[1])
					}
					cardID = parsedURL.ParentID
				}
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
				if parsedURL.AccountID > 0 {
					accountID = strconv.FormatInt(parsedURL.AccountID, 10)
				}
				if parsedURL.ProjectID > 0 {
					projectID = strconv.FormatInt(parsedURL.ProjectID, 10)
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

			// Mark step as completed
			err = client.SetStepCompletion(ctx, projectID, stepID, true)
			if err != nil {
				return fmt.Errorf("failed to mark step as completed: %w", err)
			}

			// Display success message
			fmt.Printf("✓ Marked step %d as completed in card #%d\n", stepID, cardID)

			// TODO: If note flag is provided, add a comment to the card
			if note != "" {
				fmt.Printf("Note: Adding comments to cards is not yet implemented\n")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")
	cmd.Flags().StringVar(&note, "note", "", "Add a completion note")

	return cmd
}
