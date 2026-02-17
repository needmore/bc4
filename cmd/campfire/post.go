package campfire

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/markdown"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
)

func newPostCmd(f *factory.Factory) *cobra.Command {
	var campfireFlag string

	cmd := &cobra.Command{
		Use:   "post <message>",
		Short: "Post a message to a campfire",
		Long:  `Post a message to a campfire. The message is required.`,
		Args:  cobra.ExactArgs(1),
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

			// Determine which campfire to post to
			var campfireID int64
			var campfire *api.Campfire

			if campfireFlag != "" {
				// Check if it's a URL
				if parser.IsBasecampURL(campfireFlag) {
					parsed, err := parser.ParseBasecampURL(campfireFlag)
					if err != nil {
						return fmt.Errorf("invalid Basecamp URL: %w", err)
					}
					if parsed.ResourceType != parser.ResourceTypeCampfire {
						return fmt.Errorf("URL is not a campfire URL: %s", campfireFlag)
					}
					campfireID = parsed.ResourceID
				} else {
					// Flag overrides default
					id, err := strconv.ParseInt(campfireFlag, 10, 64)
					if err == nil {
						campfireID = id
					} else {
						// It's a name, find by name
						cf, err := campfireOps.GetCampfireByName(f.Context(), projectID, campfireFlag)
						if err != nil {
							return fmt.Errorf("campfire '%s' not found", campfireFlag)
						}
						campfireID = cf.ID
						campfire = cf
					}
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
				cf, err := campfireOps.GetCampfire(f.Context(), projectID, campfireID)
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

			converter := markdown.NewConverter()
			richContent, err := converter.MarkdownToRichText(content)
			if err != nil {
				return fmt.Errorf("failed to convert message: %w", err)
			}

			// Post the message
			line, err := campfireOps.PostCampfireLine(f.Context(), projectID, campfireID, richContent, "text/html")
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
	cmd.Flags().StringVarP(&campfireFlag, "campfire", "c", "", "Campfire to post to (ID, name, or URL)")

	return cmd
}

// Helper function to check if stdout is a terminal
func isTerminal() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
