package campfire

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/needmore/bc4/internal/utils"
	"github.com/spf13/cobra"
)

func newViewCmd(f *factory.Factory) *cobra.Command {
	var limit int
	var noPager bool

	cmd := &cobra.Command{
		Use:   "view [ID|name|URL]",
		Short: "View recent messages in a campfire",
		Long: `Display recent messages from a campfire. If no campfire is specified, uses the default campfire.

You can specify the campfire using:
- A numeric ID (e.g., "12345")
- A campfire name (e.g., "General")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/chats/12345")`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get required dependencies
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			accountID, err := f.AccountID()
			if err != nil {
				return err
			}

			projectID, err := f.ProjectID()
			if err != nil {
				return err
			}

			client, err := f.ApiClient()
			if err != nil {
				return err
			}
			campfireOps := client.Campfires()

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
				// Try to parse as URL first
				if parser.IsBasecampURL(args[0]) {
					parsedURL, err := parser.ParseBasecampURL(args[0])
					if err != nil {
						return fmt.Errorf("invalid Basecamp URL: %s", args[0])
					}
					if parsedURL.ResourceType != parser.ResourceTypeCampfire {
						return fmt.Errorf("URL is not for a campfire: %s", args[0])
					}
					campfireID = parsedURL.ResourceID
					// Override account and project IDs if provided in URL
					if parsedURL.AccountID > 0 {
						accountID = strconv.FormatInt(parsedURL.AccountID, 10)
					}
					if parsedURL.ProjectID > 0 {
						projectID = strconv.FormatInt(parsedURL.ProjectID, 10)
					}
					// Need to recreate the client with the new account ID
					f = f.WithAccount(accountID).WithProject(projectID)
					client, err = f.ApiClient()
					if err != nil {
						return err
					}
					campfireOps = client.Campfires()
				} else {
					// Try to parse as ID
					id, err := strconv.ParseInt(args[0], 10, 64)
					if err == nil {
						campfireID = id
					} else {
						// It's a name, find by name
						cf, err := campfireOps.GetCampfireByName(f.Context(), projectID, args[0])
						if err != nil {
							return fmt.Errorf("campfire '%s' not found", args[0])
						}
						campfireID = cf.ID
						campfire = cf
					}
				}
			}

			// Get campfire details if we don't have them yet
			if campfire == nil {
				cf, err := campfireOps.GetCampfire(f.Context(), projectID, campfireID)
				if err != nil {
					return fmt.Errorf("failed to get campfire: %w", err)
				}
				campfire = cf
			}

			// Get campfire lines
			lines, err := campfireOps.GetCampfireLines(f.Context(), projectID, campfireID, limit)
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
