package todo

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
)

func newSetCmd() *cobra.Command {
	var accountID string
	var projectID string

	cmd := &cobra.Command{
		Use:   "set [todo-list-id]",
		Short: "Set default todo list",
		Long:  `Set the default todo list for the current project. Use 'todo select' for interactive selection.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			todoListID := args[0]

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

			// Use specified account or default
			if accountID == "" {
				accountID = authClient.GetDefaultAccount()
			}

			if accountID == "" {
				return fmt.Errorf("no account specified and no default account set")
			}

			// Use specified project or default
			if projectID == "" {
				projectID = cfg.DefaultProject
				if projectID == "" && cfg.Accounts != nil {
					if acc, ok := cfg.Accounts[accountID]; ok {
						projectID = acc.DefaultProject
					}
				}
			}

			if projectID == "" {
				return fmt.Errorf("no project specified and no default project set")
			}

			// Update config
			if cfg.Accounts == nil {
				cfg.Accounts = make(map[string]config.AccountConfig)
			}

			acc := cfg.Accounts[accountID]
			if acc.ProjectDefaults == nil {
				acc.ProjectDefaults = make(map[string]config.ProjectDefaults)
			}

			projDefaults := acc.ProjectDefaults[projectID]
			projDefaults.DefaultTodoList = todoListID
			acc.ProjectDefaults[projectID] = projDefaults
			cfg.Accounts[accountID] = acc

			// Save config
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("Default todo list set to %s for project %s\n", todoListID, projectID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID (overrides default)")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID (overrides default)")

	return cmd
}

