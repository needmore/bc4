package campfire

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
	"github.com/spf13/cobra"
)

func newSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set [ID|name]",
		Short: "Set the default campfire for the current project",
		Long:  `Set the default campfire by ID or name. This campfire will be used when no specific campfire is specified in commands.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Check if we have auth
			if cfg.ClientID == "" || cfg.ClientSecret == "" {
				return fmt.Errorf("not authenticated. Run 'bc4' to set up authentication")
			}

			// Create auth client
			authClient := auth.NewClient(cfg.ClientID, cfg.ClientSecret)

			// Use default account
			accountID := authClient.GetDefaultAccount()
			if accountID == "" {
				return fmt.Errorf("no default account set. Use 'bc4 account select' to set one")
			}

			// Get project ID
			projectID := cfg.DefaultProject
			if cfg.Accounts != nil && cfg.Accounts[accountID].DefaultProject != "" {
				projectID = cfg.Accounts[accountID].DefaultProject
			}
			if projectID == "" {
				return fmt.Errorf("no default project set. Use 'bc4 project select' to set one")
			}
			// Get token
			token, err := authClient.GetToken(accountID)
			if err != nil {
				return fmt.Errorf("failed to get auth token: %w", err)
			}

			// Create API client
			client := api.NewModularClient(accountID, token.AccessToken)
			campfireOps := client.Campfires()

			// Parse campfire ID or name
			campfireArg := args[0]
			var campfireID int64
			var campfireName string

			// Try to parse as ID first
			id, err := strconv.ParseInt(campfireArg, 10, 64)
			if err == nil {
				// It's an ID
				campfireID = id
				// Verify it exists
				campfire, err := campfireOps.GetCampfire(context.Background(), projectID, campfireID)
				if err != nil {
					return fmt.Errorf("campfire with ID %d not found", campfireID)
				}
				campfireName = campfire.Name
			} else {
				// It's a name, find by name
				campfire, err := campfireOps.GetCampfireByName(context.Background(), projectID, campfireArg)
				if err != nil {
					return fmt.Errorf("campfire '%s' not found", campfireArg)
				}
				campfireID = campfire.ID
				campfireName = campfire.Name
			}

			// Initialize accounts map if needed
			if cfg.Accounts == nil {
				cfg.Accounts = make(map[string]config.AccountConfig)
			}

			// Get or create account config
			acc := cfg.Accounts[accountID]

			// Initialize project defaults if needed
			if acc.ProjectDefaults == nil {
				acc.ProjectDefaults = make(map[string]config.ProjectDefaults)
			}

			// Update default campfire
			projDefaults := acc.ProjectDefaults[projectID]
			projDefaults.DefaultCampfire = strconv.FormatInt(campfireID, 10)
			acc.ProjectDefaults[projectID] = projDefaults
			cfg.Accounts[accountID] = acc

			// Save config
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			// Display success message
			if campfireName != "" {
				fmt.Fprintf(os.Stderr, "✓ Set default campfire to: %s (#%d)\n", campfireName, campfireID)
			} else {
				fmt.Fprintf(os.Stderr, "✓ Set default campfire to: #%d\n", campfireID)
			}

			return nil
		},
	}

	return cmd
}
