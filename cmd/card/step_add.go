package card

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/needmore/bc4/internal/utils"
	"github.com/spf13/cobra"
)

// newStepAddCmd creates the step add command
func newStepAddCmd(f *factory.Factory) *cobra.Command {
	var accountID string
	var projectID string

	cmd := &cobra.Command{
		Use:   "add [CARD_ID or URL] \"Step content\"",
		Short: "Add a new step to a card",
		Long: `Add a new step (subtask) to an existing card.

You can specify the card using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345")

Examples:
  bc4 card step add 123 "Review PR comments"
  bc4 card step add 456 "Update documentation" --assign @jane
  bc4 card step add https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345 "Fix bug"`,
		Args: cobra.ExactArgs(2),
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

			// Get API client from factory
			client, err := f.ApiClient()
			if err != nil {
				return err
			}

			// Get resolved project ID
			resolvedProjectID, err := f.ProjectID()
			if err != nil {
				return err
			}

			// Create the step request
			req := api.StepCreateRequest{
				Title: args[1],
			}

			// Handle assignees if specified
			assignFlag, _ := cmd.Flags().GetString("assign")
			if assignFlag != "" {
				// Create user resolver
				userResolver := utils.NewUserResolver(client.Client, resolvedProjectID)

				// Parse comma-separated assignees
				assignees := strings.Split(assignFlag, ",")
				for i := range assignees {
					assignees[i] = strings.TrimSpace(assignees[i])
				}

				// Resolve user identifiers to person IDs
				personIDs, err := userResolver.ResolveUsers(f.Context(), assignees)
				if err != nil {
					return fmt.Errorf("failed to resolve assignees: %w", err)
				}

				// Convert person IDs to comma-separated string
				var idStrings []string
				for _, id := range personIDs {
					idStrings = append(idStrings, strconv.FormatInt(id, 10))
				}
				req.Assignees = strings.Join(idStrings, ",")
			}

			// Create the step
			step, err := client.Steps().CreateStep(f.Context(), resolvedProjectID, cardID, req)
			if err != nil {
				return fmt.Errorf("failed to create step: %w", err)
			}

			// Output the step ID
			fmt.Printf("#%d\n", step.ID)
			return nil
		},
	}

	// TODO: Add flags for assignee, position, etc.
	cmd.Flags().String("assign", "", "Assign the step to a user")
	cmd.Flags().String("after", "", "Position step after another step ID")
	cmd.Flags().String("before", "", "Position step before another step ID")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")

	return cmd
}
