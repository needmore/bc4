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

func newSetCmd() *cobra.Command {
	var accountID string
	var projectID string

	cmd := &cobra.Command{
		Use:   "set [ID|name]",
		Short: "Set default card table",
		Long:  `Set the default card table for the current project. This default will be used when no table is specified in other commands.`,
		Args:  cobra.ExactArgs(1),
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
			client := api.NewClient(accountID, token.AccessToken)

			// Parse input as ID or search by name
			var cardTableID int64
			input := args[0]

			// Try to parse as ID first
			if id, err := strconv.ParseInt(input, 10, 64); err == nil {
				cardTableID = id
				// Verify it exists
				if _, err := client.GetCardTable(ctx, projectID, cardTableID); err != nil {
					return fmt.Errorf("card table %d not found", cardTableID)
				}
			} else {
				// Search by name - for now just get the project's card table
				cardTable, err := client.GetProjectCardTable(ctx, projectID)
				if err != nil {
					return fmt.Errorf("failed to fetch card table: %w", err)
				}

				// Check if name matches
				if strings.Contains(strings.ToLower(cardTable.Title), strings.ToLower(input)) {
					cardTableID = cardTable.ID
				} else {
					return fmt.Errorf("no card table found matching '%s'", input)
				}
			}

			// Initialize project defaults if needed
			if cfg.Accounts == nil {
				cfg.Accounts = make(map[string]config.AccountConfig)
			}
			if _, ok := cfg.Accounts[accountID]; !ok {
				cfg.Accounts[accountID] = config.AccountConfig{
					ProjectDefaults: make(map[string]config.ProjectDefaults),
				}
			}
			if cfg.Accounts[accountID].ProjectDefaults == nil {
				acc := cfg.Accounts[accountID]
				acc.ProjectDefaults = make(map[string]config.ProjectDefaults)
				cfg.Accounts[accountID] = acc
			}

			// Set the default card table
			acc := cfg.Accounts[accountID]
			proj := acc.ProjectDefaults[projectID]
			proj.DefaultCardTable = fmt.Sprintf("%d", cardTableID)
			acc.ProjectDefaults[projectID] = proj
			cfg.Accounts[accountID] = acc

			// Save config
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("Set default card table to %d\n", cardTableID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")

	return cmd
}
