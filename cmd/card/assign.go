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

func newAssignCmd(f *factory.Factory) *cobra.Command {
	var accountID string
	var projectID string
	var assign []string

	cmd := &cobra.Command{
		Use:   "assign [ID or URL] [@user ...]",
		Short: "Assign people to card",
		Long: `Assign one or more people to a card.

You can specify the card using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345")

Users can be specified as positional arguments or with the --assign flag.
Supports email addresses, @mentions, and names.`,
		Example: `  # Assign by @mention
  bc4 card assign 12345 @jane

  # Assign multiple users
  bc4 card assign 12345 @jane @john

  # Assign by email
  bc4 card assign 12345 --assign user@example.com

  # Assign using a URL
  bc4 card assign https://3.basecamp.com/.../cards/12345 @jane`,
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

			// Collect user identifiers from positional args and --assign flag
			userIdentifiers := append(args[1:], assign...)
			if len(userIdentifiers) == 0 {
				return fmt.Errorf("no users specified. Provide users as arguments or with --assign")
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

			// Get the card to merge with existing assignees
			card, err := cardOps.GetCard(f.Context(), resolvedProjectID, cardID)
			if err != nil {
				return fmt.Errorf("failed to fetch card: %w", err)
			}

			// Resolve user identifiers to person IDs
			userResolver := utils.NewUserResolver(client.Client, resolvedProjectID)
			newAssigneeIDs, err := userResolver.ResolveUsers(f.Context(), userIdentifiers)
			if err != nil {
				return fmt.Errorf("failed to resolve assignees: %w", err)
			}

			// Start with existing assignees
			currentAssigneeIDs := make([]int64, 0, len(card.Assignees))
			for _, assignee := range card.Assignees {
				currentAssigneeIDs = append(currentAssigneeIDs, assignee.ID)
			}

			// Merge without duplicates
			for _, newID := range newAssigneeIDs {
				found := false
				for _, existingID := range currentAssigneeIDs {
					if existingID == newID {
						found = true
						break
					}
				}
				if !found {
					currentAssigneeIDs = append(currentAssigneeIDs, newID)
				}
			}

			// Update the card
			req := api.CardUpdateRequest{
				Title:       card.Title,
				Content:     card.Content,
				AssigneeIDs: currentAssigneeIDs,
			}
			_, err = cardOps.UpdateCard(f.Context(), resolvedProjectID, cardID, req)
			if err != nil {
				return fmt.Errorf("failed to assign users: %w", err)
			}

			fmt.Printf("Assigned %d user(s) to card #%d\n", len(newAssigneeIDs), card.ID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")
	cmd.Flags().StringSliceVar(&assign, "assign", nil, "Add assignees (by email or @mention)")

	return cmd
}
