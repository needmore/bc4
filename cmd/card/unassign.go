package card

import (
	"fmt"
	"strconv"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/needmore/bc4/internal/utils"
	"github.com/spf13/cobra"
)

func newUnassignCmd(f *factory.Factory) *cobra.Command {
	var accountID string
	var projectID string
	var unassign []string

	cmd := &cobra.Command{
		Use:   "unassign [ID or URL] [@user ...]",
		Short: "Remove assignees from card",
		Long: `Remove one or more assignees from a card.

You can specify the card using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345")

Users can be specified as positional arguments or with the --unassign flag.
Supports email addresses, @mentions, and names.`,
		Example: `  # Remove by @mention
  bc4 card unassign 12345 @jane

  # Remove multiple users
  bc4 card unassign 12345 @jane @john

  # Remove by email
  bc4 card unassign 12345 --unassign user@example.com

  # Remove using a URL
  bc4 card unassign https://3.basecamp.com/.../cards/12345 @jane`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse card ID (could be numeric ID or URL)
			cardID, parsedURL, err := parser.ParseArgument(args[0])
			if err != nil {
				return fmt.Errorf("invalid card ID or URL: %s", args[0])
			}

			// Apply overrides if specified
			if accountID != "" {
				f = f.WithAccount(accountID)
			}
			if projectID != "" {
				f = f.WithProject(projectID)
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

			// Collect user identifiers from positional args and --unassign flag
			userIdentifiers := append(args[1:], unassign...)
			if len(userIdentifiers) == 0 {
				return fmt.Errorf("no users specified. Provide users as arguments or with --unassign")
			}

			// Get resolved project ID
			resolvedProjectID, err := f.ProjectID()
			if err != nil {
				return err
			}

			// Get API client from factory
			client, err := f.ApiClient()
			if err != nil {
				return err
			}
			cardOps := client.Cards()

			// Get the card to check current assignees
			card, err := cardOps.GetCard(f.Context(), resolvedProjectID, cardID)
			if err != nil {
				return fmt.Errorf("failed to fetch card: %w", err)
			}

			if len(card.Assignees) == 0 {
				fmt.Println("Card has no assignees")
				return nil
			}

			// Resolve user identifiers to person IDs
			userResolver := utils.NewUserResolver(client.Client, resolvedProjectID)
			removeIDs, err := userResolver.ResolveUsers(f.Context(), userIdentifiers)
			if err != nil {
				return fmt.Errorf("failed to resolve assignees: %w", err)
			}

			// Filter out removed assignees
			filteredIDs := make([]int64, 0)
			removedCount := 0
			for _, assignee := range card.Assignees {
				shouldRemove := false
				for _, removeID := range removeIDs {
					if assignee.ID == removeID {
						shouldRemove = true
						break
					}
				}
				if shouldRemove {
					removedCount++
				} else {
					filteredIDs = append(filteredIDs, assignee.ID)
				}
			}

			if removedCount == 0 {
				fmt.Println("None of the specified users are currently assigned to this card")
				return nil
			}

			// Update the card
			req := api.CardUpdateRequest{
				Title:       card.Title,
				Content:     card.Content,
				AssigneeIDs: filteredIDs,
			}
			_, err = cardOps.UpdateCard(f.Context(), resolvedProjectID, cardID, req)
			if err != nil {
				return fmt.Errorf("failed to unassign users: %w", err)
			}

			fmt.Printf("Removed %d user(s) from card #%d\n", removedCount, card.ID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")
	cmd.Flags().StringSliceVar(&unassign, "unassign", nil, "Remove assignees (by email or @mention)")

	return cmd
}
