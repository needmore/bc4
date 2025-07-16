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

// newStepAssignCmd creates the step assign command
func newStepAssignCmd(f *factory.Factory) *cobra.Command {
	var accountID string
	var projectID string

	cmd := &cobra.Command{
		Use:   "assign [CARD_ID or URL] [STEP_ID or URL] USER",
		Short: "Assign a step to a user",
		Long: `Assign a step (subtask) to a specific user.

You can specify the card and step using either:
- Numeric IDs (e.g., "123 456" for card 123, step 456)
- A Basecamp step URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345/steps/67890")

Examples:
  bc4 card step assign 123 456 @jane
  bc4 card step assign 123 456 "Jane Doe"
  bc4 card step assign 123 456 jane@example.com
  bc4 card step assign https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345/steps/67890 @jane`,
		Args: cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			var cardID, stepID int64
			var parsedURL *parser.ParsedURL
			var userIdentifier string

			// Parse arguments - could be card ID + step ID + user, or step URL + user
			if len(args) == 2 {
				// Two arguments - step URL + user
				var err error
				stepID, parsedURL, err = parser.ParseArgument(args[0])
				if err != nil {
					return fmt.Errorf("invalid step URL: %s", args[0])
				}
				if parsedURL == nil {
					return fmt.Errorf("when providing two arguments, the first must be a Basecamp step URL")
				}
				if parsedURL.ResourceType != parser.ResourceTypeStep {
					return fmt.Errorf("URL is not for a step: %s", args[0])
				}
				cardID = parsedURL.ParentID
				userIdentifier = args[1]
			} else {
				// Three arguments - card ID, step ID, and user
				var err error
				cardID, _, err = parser.ParseArgument(args[0])
				if err != nil {
					return fmt.Errorf("invalid card ID or URL: %s", args[0])
				}

				stepID, parsedURL, err = parser.ParseArgument(args[1])
				if err != nil {
					return fmt.Errorf("invalid step ID or URL: %s", args[1])
				}

				// If step was provided as URL, validate and extract IDs
				if parsedURL != nil {
					if parsedURL.ResourceType != parser.ResourceTypeStep {
						return fmt.Errorf("URL is not for a step: %s", args[1])
					}
					cardID = parsedURL.ParentID
				}
				userIdentifier = args[2]
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

			// Check if we're unassigning
			unassign, _ := cmd.Flags().GetBool("unassign")

			// Get current step to preserve existing data
			card, err := client.Cards().GetCard(f.Context(), resolvedProjectID, cardID)
			if err != nil {
				return fmt.Errorf("failed to get card: %w", err)
			}

			var currentStep *api.Step
			for _, step := range card.Steps {
				if step.ID == stepID {
					currentStep = &step
					break
				}
			}

			if currentStep == nil {
				return fmt.Errorf("step with ID %d not found in card", stepID)
			}

			var assignees string
			if unassign {
				// Clear assignees
				assignees = ""
			} else {
				// Resolve user
				userResolver := utils.NewUserResolver(client.Client, resolvedProjectID)
				personIDs, err := userResolver.ResolveUsers(f.Context(), []string{userIdentifier})
				if err != nil {
					return fmt.Errorf("failed to resolve user: %w", err)
				}

				// Convert person IDs to comma-separated string
				var idStrings []string
				for _, id := range personIDs {
					idStrings = append(idStrings, strconv.FormatInt(id, 10))
				}
				assignees = strings.Join(idStrings, ",")
			}

			// Update the step
			req := api.StepUpdateRequest{
				Title:     currentStep.Title, // Preserve title
				Assignees: assignees,
			}

			_, err = client.Steps().UpdateStep(f.Context(), resolvedProjectID, stepID, req)
			if err != nil {
				return fmt.Errorf("failed to update step: %w", err)
			}

			if unassign {
				fmt.Printf("Step #%d unassigned\n", stepID)
			} else {
				fmt.Printf("Step #%d assigned to %s\n", stepID, userIdentifier)
			}
			return nil
		},
	}

	// TODO: Add flags for unassigning
	cmd.Flags().Bool("unassign", false, "Remove current assignee instead")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")

	return cmd
}
