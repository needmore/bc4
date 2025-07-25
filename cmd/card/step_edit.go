package card

import (
	"fmt"
	"strconv"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
)

// newStepEditCmd creates the step edit command
func newStepEditCmd(f *factory.Factory) *cobra.Command {
	var accountID string
	var projectID string

	cmd := &cobra.Command{
		Use:   "edit [CARD_ID or URL] [STEP_ID or URL]",
		Short: "Edit a step's content",
		Long: `Edit the content of an existing step (subtask).

You can specify the card and step using either:
- Numeric IDs (e.g., "123 456" for card 123, step 456)
- A Basecamp step URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345/steps/67890")

Examples:
  bc4 card step edit 123 456 --content "Updated step description"
  bc4 card step edit 123 456 --interactive
  bc4 card step edit https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345/steps/67890 --content "New content"`,
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

			// Get content from flag or interactive editor
			content, _ := cmd.Flags().GetString("content")
			interactive, _ := cmd.Flags().GetBool("interactive")

			if content == "" && !interactive {
				// Fetch current step content
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

				// For now, just use the current content
				// TODO: Implement editor integration
				return fmt.Errorf("interactive editing not yet implemented - use --content flag")
			} else if interactive {
				// For now, just error
				// TODO: Implement editor integration
				return fmt.Errorf("interactive editing not yet implemented - use --content flag")
			}

			if content == "" {
				return fmt.Errorf("no content provided")
			}

			// Update the step
			req := api.StepUpdateRequest{
				Title: content,
			}

			_, err = client.Steps().UpdateStep(f.Context(), resolvedProjectID, stepID, req)
			if err != nil {
				return fmt.Errorf("failed to update step: %w", err)
			}

			fmt.Printf("Step #%d updated\n", stepID)
			return nil
		},
	}

	// TODO: Add flags for content and interactive mode
	cmd.Flags().String("content", "", "New content for the step")
	cmd.Flags().Bool("interactive", false, "Open interactive editor")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")

	return cmd
}
