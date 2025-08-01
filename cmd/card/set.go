package card

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/needmore/bc4/internal/config"
	"github.com/needmore/bc4/internal/factory"
	"github.com/spf13/cobra"
)

func newSetCmd(f *factory.Factory) *cobra.Command {
	var accountID string
	var projectID string

	cmd := &cobra.Command{
		Use:   "set [ID|name]",
		Short: "Set default card table",
		Long:  `Set the default card table for the current project. This default will be used when no table is specified in other commands.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Apply overrides if specified
			f = f.ApplyOverrides(accountID, projectID)

			// Get API client from factory
			client, err := f.ApiClient()
			if err != nil {
				return err
			}
			cardOps := client.Cards()

			// Get resolved IDs
			resolvedAccountID, err := f.AccountID()
			if err != nil {
				return err
			}
			resolvedProjectID, err := f.ProjectID()
			if err != nil {
				return err
			}

			// Get config for updating
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			// Parse input as ID or search by name
			var cardTableID int64
			input := args[0]

			// Try to parse as ID first
			if id, err := strconv.ParseInt(input, 10, 64); err == nil {
				cardTableID = id
				// Verify it exists
				if _, err := cardOps.GetCardTable(f.Context(), resolvedProjectID, cardTableID); err != nil {
					return fmt.Errorf("card table %d not found", cardTableID)
				}
			} else {
				// Search by name - for now just get the project's card table
				cardTable, err := cardOps.GetProjectCardTable(f.Context(), resolvedProjectID)
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
			if _, ok := cfg.Accounts[resolvedAccountID]; !ok {
				cfg.Accounts[resolvedAccountID] = config.AccountConfig{
					ProjectDefaults: make(map[string]config.ProjectDefaults),
				}
			}
			if cfg.Accounts[resolvedAccountID].ProjectDefaults == nil {
				acc := cfg.Accounts[resolvedAccountID]
				acc.ProjectDefaults = make(map[string]config.ProjectDefaults)
				cfg.Accounts[resolvedAccountID] = acc
			}

			// Set the default card table
			acc := cfg.Accounts[resolvedAccountID]
			proj := acc.ProjectDefaults[resolvedProjectID]
			proj.DefaultCardTable = fmt.Sprintf("%d", cardTableID)
			acc.ProjectDefaults[resolvedProjectID] = proj
			cfg.Accounts[resolvedAccountID] = acc

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
