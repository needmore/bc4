package campfire

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
	"github.com/spf13/cobra"
)

func newPostCmd() *cobra.Command {
	var campfireFlag string

	cmd := &cobra.Command{
		Use:   "post <message>",
		Short: "Post a message to a campfire",
		Long:  `Post a message to a campfire. The message is required.`,
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
			client := api.NewClient(accountID, token.AccessToken)

			// Determine which campfire to post to
			var campfireID int64
			var campfire *api.Campfire

			if campfireFlag != "" {
				// Flag overrides default
				id, err := strconv.ParseInt(campfireFlag, 10, 64)
				if err == nil {
					campfireID = id
				} else {
					// It's a name, find by name
					cf, err := client.GetCampfireByName(context.Background(), projectID, campfireFlag)
					if err != nil {
						return fmt.Errorf("campfire '%s' not found", campfireFlag)
					}
					campfireID = cf.ID
					campfire = cf
				}
			} else {
				// Use default campfire
				defaultCampfireID := ""
				if cfg.Accounts != nil && cfg.Accounts[accountID].ProjectDefaults != nil {
					if projDefaults, ok := cfg.Accounts[accountID].ProjectDefaults[projectID]; ok {
						defaultCampfireID = projDefaults.DefaultCampfire
					}
				}
				if defaultCampfireID == "" {
					return fmt.Errorf("no campfire specified and no default set. Use 'campfire set' to set a default or use --campfire flag")
				}
				campfireID, _ = strconv.ParseInt(defaultCampfireID, 10, 64)
			}

			// Get campfire details if we don't have them yet
			if campfire == nil {
				cf, err := client.GetCampfire(context.Background(), projectID, campfireID)
				if err != nil {
					return fmt.Errorf("failed to get campfire: %w", err)
				}
				campfire = cf
			}

			// Get message content
			content := args[0]

			// Trim whitespace
			content = strings.TrimSpace(content)
			if content == "" {
				return fmt.Errorf("message cannot be empty")
			}

			// Post the message
			line, err := client.PostCampfireLine(context.Background(), projectID, campfireID, content)
			if err != nil {
				return fmt.Errorf("failed to post message: %w", err)
			}

			// Success message
			fmt.Fprintf(os.Stderr, "✓ Posted to %s\n", campfire.Name)

			// In non-TTY mode, output the line ID
			if !isTerminal() {
				fmt.Println(line.ID)
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&campfireFlag, "campfire", "c", "", "Campfire to post to (ID or name)")

	return cmd
}

// Helper function to check if stdout is a terminal
func isTerminal() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
