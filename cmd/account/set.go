package account

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
)

func newSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set [account-id]",
		Short: "Set default account",
		Long:  `Set the default account for bc4 commands. Use 'account select' for interactive selection.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			accountID := args[0]

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

			// Verify the account exists
			accounts := authClient.GetAccounts()
			account, exists := accounts[accountID]
			if !exists {
				return fmt.Errorf("account %s not found", accountID)
			}

			// Check if we're changing accounts
			oldDefaultAccount := authClient.GetDefaultAccount()
			changingAccounts := oldDefaultAccount != "" && oldDefaultAccount != accountID

			// Set default account
			if err := authClient.SetDefaultAccount(accountID); err != nil {
				return fmt.Errorf("failed to set default account: %w", err)
			}

			// Update config
			cfg.DefaultAccount = accountID
			
			// Clear default project if changing accounts
			if changingAccounts {
				cfg.DefaultProject = ""
				// Also clear the account-specific default project
				if cfg.Accounts != nil {
					for accID, accConfig := range cfg.Accounts {
						if accID == accountID {
							accConfig.DefaultProject = ""
							cfg.Accounts[accID] = accConfig
						}
					}
				}
			}
			
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("Default account set to: %s (ID: %s)\n", account.AccountName, accountID)
			if changingAccounts {
				fmt.Println("Note: Default project has been cleared since you changed accounts.")
			}
			return nil
		},
	}

	return cmd
}