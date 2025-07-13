package card

import (
	"context"
	"fmt"
	"strconv"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
	"github.com/spf13/cobra"
)

// newStepUncheckCmd creates the step uncheck command
func newStepUncheckCmd() *cobra.Command {
	var accountID string
	var projectID string
	var reason string

	cmd := &cobra.Command{
		Use:   "uncheck CARD_ID STEP_ID",
		Short: "Mark a step as incomplete",
		Long: `Mark a completed step (subtask) as incomplete again.

Examples:
  bc4 card step uncheck 123 456
  bc4 card step uncheck 123 456 --reason "Needs rework"`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Parse card ID and step ID
			cardID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid card ID: %s", args[0])
			}

			stepID, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid step ID: %s", args[1])
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

			// Mark step as incomplete
			err = client.SetStepCompletion(ctx, projectID, stepID, false)
			if err != nil {
				return fmt.Errorf("failed to mark step as incomplete: %w", err)
			}

			// Display success message
			fmt.Printf("○ Marked step %d as incomplete in card #%d\n", stepID, cardID)

			// TODO: If reason flag is provided, add a comment to the card
			if reason != "" {
				fmt.Printf("Note: Adding comments to cards is not yet implemented\n")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")
	cmd.Flags().StringVar(&reason, "reason", "", "Add a reason for marking incomplete")

	return cmd
}
