package card

import (
	"fmt"
	"strconv"

	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
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

			// Get API client from factory (for auth check)
			_, err := f.ApiClient()
			if err != nil {
				return err
			}

			// TODO: Implement step assign functionality
			// 4. Resolve user (by handle, name, or email)
			// 5. Call API to assign step
			// 6. Display success message
			fmt.Printf("Would assign step %d in card #%d to user: %s\n", stepID, cardID, userIdentifier)
			return fmt.Errorf("step assign not yet implemented")
		},
	}

	// TODO: Add flags for unassigning
	cmd.Flags().Bool("unassign", false, "Remove current assignee instead")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")

	return cmd
}
