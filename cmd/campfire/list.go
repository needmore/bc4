package campfire

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
	"github.com/needmore/bc4/internal/ui/tableprinter"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all campfires in the project",
		Long:  `Display all campfires (chat rooms) in the current project with their status and activity.`,
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
			client := api.NewClient(accountID, token.AccessToken)

			// Get campfires
			campfires, err := client.ListCampfires(context.Background(), projectID)
			if err != nil {
				return fmt.Errorf("failed to list campfires: %w", err)
			}

			if len(campfires) == 0 {
				fmt.Println("No campfires found in this project.")
				return nil
			}

			// Get default campfire ID from config
			defaultCampfireID := ""
			if cfg.Accounts != nil && cfg.Accounts[accountID].ProjectDefaults != nil {
				if projDefaults, ok := cfg.Accounts[accountID].ProjectDefaults[projectID]; ok {
					defaultCampfireID = projDefaults.DefaultCampfire
				}
			}

			// Create table
			table := tableprinter.New(os.Stdout)

			// Add headers
			if table.IsTTY() {
				table.AddHeader("ID", "NAME", "STATUS", "LAST ACTIVITY")
			} else {
				table.AddHeader("ID", "NAME", "STATUS", "STATE", "LAST_ACTIVITY")
			}

			// Sort campfires by updated_at (most recent first)
			for i := len(campfires) - 1; i >= 0; i-- {
				cf := campfires[i]
				idStr := strconv.FormatInt(cf.ID, 10)

				// Add ID field with default indicator
				if idStr == defaultCampfireID {
					if table.IsTTY() {
						table.AddIDField(idStr+"*", cf.Status) // Add asterisk for default
					} else {
						table.AddField(idStr)
					}
				} else {
					table.AddIDField(idStr, cf.Status)
				}

				// Add name field
				name := cf.Name
				if name == "" {
					name = "(untitled)"
				}
				table.AddProjectField(name, cf.Status)

				// Add status field
				if table.IsTTY() {
					table.AddColorField(cf.Status, cf.Status)
				} else {
					table.AddField(cf.Status)
				}

				// Add state column for non-TTY
				if !table.IsTTY() {
					table.AddField(cf.Status)
				}

				// Add last activity
				now := time.Now()
				table.AddTimeField(now, cf.UpdatedAt)

				table.EndRow()
			}

			// Render table
			return table.Render()
		},
	}

	return cmd
}