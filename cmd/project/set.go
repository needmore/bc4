package project

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/needmore/bc4/internal/config"
	"github.com/needmore/bc4/internal/factory"
)

func newSetCmd(f *factory.Factory) *cobra.Command {
	var accountID string

	cmd := &cobra.Command{
		Use:   "set [project-id]",
		Short: "Set default project",
		Long:  `Set the default project for bc4 commands. Use 'project select' for interactive selection.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID := args[0]

			// Get config and auth through factory
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			authClient, err := f.AuthClient()
			if err != nil {
				return err
			}

			// If accountID was specified, use a new factory with that account
			if accountID != "" {
				f = f.WithAccount(accountID)
			} else {
				// Use default account
				accountID = authClient.GetDefaultAccount()
				if accountID == "" {
					return fmt.Errorf("no account specified and no default account set")
				}
			}

			// Update config
			cfg.DefaultProject = projectID

			if cfg.Accounts == nil {
				cfg.Accounts = make(map[string]config.AccountConfig)
			}

			// Update account-specific default project
			accountCfg := cfg.Accounts[accountID]
			accountCfg.DefaultProject = projectID
			// Preserve the name if it exists
			if accountCfg.Name == "" {
				// Get the account name from auth
				if token, err := authClient.GetToken(accountID); err == nil {
					accountCfg.Name = token.AccountName
				}
			}
			cfg.Accounts[accountID] = accountCfg

			// Save config
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("Default project set to: %s\n", projectID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID (overrides default)")

	return cmd
}
