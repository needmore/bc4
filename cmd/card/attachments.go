package card

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	attachmentsCmd "github.com/needmore/bc4/cmd/attachments"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
)

func newAttachmentsCmd(f *factory.Factory) *cobra.Command {
	var accountID string
	var projectID string

	cmd := &cobra.Command{
		Use:   "attachments [card-id or URL]",
		Short: "List attachments for a card",
		Long: `List all attachments (images and files) associated with a card.

You can specify the card using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345")`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Apply account override if specified
			if accountID != "" {
				f = f.WithAccount(accountID)
			}

			// Apply project override if specified
			if projectID != "" {
				f = f.WithProject(projectID)
			}

			// Parse card ID (could be numeric ID or URL)
			cardID, parsedURL, err := parser.ParseArgument(args[0])
			if err != nil {
				return fmt.Errorf("invalid card ID or URL: %s", args[0])
			}

			// If a URL was parsed, override account and project IDs if provided
			if parsedURL != nil {
				if parsedURL.ResourceType != parser.ResourceTypeCard {
					return fmt.Errorf("URL is not for a card: %s", args[0])
				}
				if parsedURL.AccountID > 0 {
					f = f.WithAccount(strconv.FormatInt(parsedURL.AccountID, 10))
				}
				if parsedURL.ProjectID > 0 {
					f = f.WithProject(strconv.FormatInt(parsedURL.ProjectID, 10))
				}
			}

			// Get API client from factory
			client, err := f.ApiClient()
			if err != nil {
				return err
			}
			cardOps := client.Cards()

			// Get resolved project ID
			resolvedProjectID, err := f.ProjectID()
			if err != nil {
				return err
			}

			// Fetch the card details
			card, err := cardOps.GetCard(f.Context(), resolvedProjectID, cardID)
			if err != nil {
				return fmt.Errorf("failed to get card: %w", err)
			}

			// Display attachments from the content
			return attachmentsCmd.DisplayAttachments(card.Content)
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")

	return cmd
}
