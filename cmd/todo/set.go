package todo

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/needmore/bc4/internal/config"
	"github.com/needmore/bc4/internal/factory"
)

func newSetCmd(f *factory.Factory) *cobra.Command {
	var accountID string
	var projectID string

	cmd := &cobra.Command{
		Use:   "set [todo-list-id]",
		Short: "Set default todo list",
		Long:  `Set the default todo list for the current project. Use 'todo select' for interactive selection.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			todoListID := args[0]

			// Apply account override if specified
			if accountID != "" {
				f = f.WithAccount(accountID)
			}

			// Apply project override if specified
			if projectID != "" {
				f = f.WithProject(projectID)
			}

			// Get config
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			// Get resolved account ID
			resolvedAccountID, err := f.AccountID()
			if err != nil {
				return err
			}

			// Get resolved project ID
			resolvedProjectID, err := f.ProjectID()
			if err != nil {
				return err
			}

			// Update config
			if cfg.Accounts == nil {
				cfg.Accounts = make(map[string]config.AccountConfig)
			}

			acc := cfg.Accounts[resolvedAccountID]
			if acc.ProjectDefaults == nil {
				acc.ProjectDefaults = make(map[string]config.ProjectDefaults)
			}

			projDefaults := acc.ProjectDefaults[resolvedProjectID]
			projDefaults.DefaultTodoList = todoListID
			acc.ProjectDefaults[resolvedProjectID] = projDefaults
			cfg.Accounts[resolvedAccountID] = acc

			// Save config
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("Default todo list set to %s for project %s\n", todoListID, resolvedProjectID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID (overrides default)")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID (overrides default)")

	return cmd
}
