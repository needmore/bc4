package card

import (
	"fmt"
	"strconv"

	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
)

// newStepUncheckCmd creates the step uncheck command
func newStepUncheckCmd(f *factory.Factory) *cobra.Command {
	var accountID string
	var projectID string
	var reason string

	cmd := &cobra.Command{
		Use:   "uncheck [CARD_ID or URL] [STEP_ID or URL]",
		Short: "Mark a step as incomplete",
		Long: `Mark a completed step (subtask) as incomplete again.

You can specify the card and step using either:
- Numeric IDs (e.g., "123 456" for card 123, step 456)
- A Basecamp step URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345/steps/67890")

Examples:
  bc4 card step uncheck 123 456
  bc4 card step uncheck 123 456 --reason "Needs rework"
  bc4 card step uncheck https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345/steps/67890`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var cardID, stepID int64
			var parsedURL *parser.ParsedURL

			// Parse arguments - could be card ID + step ID, or a single step URL
			if len(args) == 1 {
				// Single argument - must be a step URL
				var err error
				stepID, parsedURL, err = parser.ParseArgument(args[0])
				if err != nil {
					return fmt.Errorf("invalid step URL: %s", args[0])
				}
				if parsedURL == nil {
					return fmt.Errorf("when providing a single argument, it must be a Basecamp step URL")
				}
				if parsedURL.ResourceType != parser.ResourceTypeStep {
					return fmt.Errorf("URL is not for a step: %s", args[0])
				}
				cardID = parsedURL.ParentID
			} else {
				// Two arguments - card ID and step ID
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
			stepOps := client.Steps()

			// Mark step as incomplete
			err = stepOps.SetStepCompletion(f.Context(), resolvedProjectID, stepID, false)
			if err != nil {
				return fmt.Errorf("failed to mark step as incomplete: %w", err)
			}

			// Display success message
			fmt.Printf("○ Marked step %d as incomplete in card #%d\n", stepID, cardID)

			// TODO: If reason flag is provided, add a comment to the card
			if reason != "" {
				fmt.Printf("Note: Adding comments to cards is not yet implemented\n")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")
	cmd.Flags().StringVar(&reason, "reason", "", "Add a reason for marking incomplete")

	return cmd
}
