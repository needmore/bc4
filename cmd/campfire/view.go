package campfire

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
	"github.com/needmore/bc4/internal/utils"
	"github.com/spf13/cobra"
)

func newViewCmd() *cobra.Command {
	var limit int
	var noPager bool

	cmd := &cobra.Command{
		Use:   "view [ID|name]",
		Short: "View recent messages in a campfire",
		Long:  `Display recent messages from a campfire. If no campfire is specified, uses the default campfire.`,
		Args:  cobra.MaximumNArgs(1),
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

			// Determine which campfire to view
			var campfireID int64
			var campfire *api.Campfire

			if len(args) == 0 {
				// No argument - use default campfire if set
				defaultCampfireID := ""
				if cfg.Accounts != nil && cfg.Accounts[accountID].ProjectDefaults != nil {
					if projDefaults, ok := cfg.Accounts[accountID].ProjectDefaults[projectID]; ok {
						defaultCampfireID = projDefaults.DefaultCampfire
					}
				}
				if defaultCampfireID == "" {
					return fmt.Errorf("no campfire specified and no default set. Use 'campfire set' to set a default")
				}
				campfireID, _ = strconv.ParseInt(defaultCampfireID, 10, 64)
			} else {
				// Try to parse as ID first
				id, err := strconv.ParseInt(args[0], 10, 64)
				if err == nil {
					campfireID = id
				} else {
					// It's a name, find by name
					cf, err := client.GetCampfireByName(context.Background(), projectID, args[0])
					if err != nil {
						return fmt.Errorf("campfire '%s' not found", args[0])
					}
					campfireID = cf.ID
					campfire = cf
				}
			}

			// Get campfire details if we don't have them yet
			if campfire == nil {
				cf, err := client.GetCampfire(context.Background(), projectID, campfireID)
				if err != nil {
					return fmt.Errorf("failed to get campfire: %w", err)
				}
				campfire = cf
			}

			// Get campfire lines
			lines, err := client.GetCampfireLines(context.Background(), projectID, campfireID, limit)
			if err != nil {
				return fmt.Errorf("failed to get campfire lines: %w", err)
			}

			// Prepare output for pager
			var buf bytes.Buffer

			// Display header
			fmt.Fprintf(&buf, "=== %s ===\n\n", campfire.Name)

			if len(lines) == 0 {
				fmt.Fprintln(&buf, "No messages in this campfire yet.")
				// Display using pager
				pagerOpts := &utils.PagerOptions{
					Pager:   cfg.Preferences.Pager,
					NoPager: noPager,
				}
				return utils.ShowInPager(buf.String(), pagerOpts)
			}

			// Display messages in chronological order (API returns newest first, so reverse)
			for i := len(lines) - 1; i >= 0; i-- {
				line := lines[i]

				// Format timestamp
				timestamp := line.CreatedAt.Local().Format("15:04")

				// Format creator name
				creatorName := line.Creator.Name
				if creatorName == "" {
					creatorName = "Unknown"
				}

				// Format and display message
				content := strings.TrimSpace(line.Content)
				if content == "" {
					continue // Skip empty messages
				}

				// Handle multi-line messages
				contentLines := strings.Split(content, "\n")
				if len(contentLines) == 1 {
					// Single line message
					fmt.Fprintf(&buf, "[%s] @%s: %s\n", timestamp, creatorName, content)
				} else {
					// Multi-line message
					fmt.Fprintf(&buf, "[%s] @%s:\n", timestamp, creatorName)
					for _, contentLine := range contentLines {
						if contentLine != "" {
							fmt.Fprintf(&buf, "  %s\n", contentLine)
						}
					}
				}
			}

			// Add spacing and info
			fmt.Fprintln(&buf)
			if limit > 0 && len(lines) == limit {
				fmt.Fprintf(&buf, "Showing last %d messages. Use --limit to see more.\n", limit)
			} else {
				fmt.Fprintf(&buf, "Showing last %d messages.\n", len(lines))
			}

			// Display using pager
			pagerOpts := &utils.PagerOptions{
				Pager:   cfg.Preferences.Pager,
				NoPager: noPager,
			}
			return utils.ShowInPager(buf.String(), pagerOpts)
		},
	}

	// Add flags
	cmd.Flags().IntVarP(&limit, "limit", "n", 50, "Number of messages to show")
	cmd.Flags().BoolVar(&noPager, "no-pager", false, "Disable pager for output")

	return cmd
}
