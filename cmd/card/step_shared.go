package card

import (
	"context"
	"fmt"
	"strconv"

	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
)

// stepOperation defines the type of operation to perform on a step
type stepOperation int

const (
	stepOperationCheck stepOperation = iota
	stepOperationUncheck
)

// createStepCommand creates a command for checking or unchecking steps
func createStepCommand(f *factory.Factory, op stepOperation) *cobra.Command {
	var accountID string
	var projectID string
	var noteOrReason string

	// Configure command based on operation
	var use, short, long, flagName, flagUsage string
	var examples []string

	switch op {
	case stepOperationCheck:
		use = "check [CARD_ID or URL] [STEP_ID or URL]"
		short = "Mark a step as completed"
		long = `Mark a step (subtask) as completed.`
		flagName = "note"
		flagUsage = "Note to include with the completion"
		examples = []string{
			"  bc4 card step check 123 456",
			"  bc4 card step check 123 456 --note \"Fixed in PR #789\"",
			"  bc4 card step check https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345/steps/67890",
		}
	case stepOperationUncheck:
		use = "uncheck [CARD_ID or URL] [STEP_ID or URL]"
		short = "Mark a step as incomplete"
		long = `Mark a completed step (subtask) as incomplete again.`
		flagName = "reason"
		flagUsage = "Reason for marking incomplete"
		examples = []string{
			"  bc4 card step uncheck 123 456",
			"  bc4 card step uncheck 123 456 --reason \"Needs rework\"",
			"  bc4 card step uncheck https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345/steps/67890",
		}
	}

	long += `

You can specify the card and step using either:
- Numeric IDs (e.g., "123 456" for card 123, step 456)
- A Basecamp step URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345/steps/67890")

Examples:
` + examples[0] + "\n" + examples[1] + "\n" + examples[2]

	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
		Args:  cobra.RangeArgs(1, 2),
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

			// Get account ID
			accountID, err := f.AccountID()
			if err != nil {
				return err
			}

			// Get project ID
			project, err := f.ProjectID()
			if err != nil {
				return err
			}

			client, err := f.ApiClient()
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Perform the operation
			completed := op == stepOperationCheck
			err = client.SetStepCompletion(ctx, project, stepID, completed)

			if err != nil {
				return fmt.Errorf("failed to update step: %w", err)
			}

			// Success message
			action := "completed"
			if op == stepOperationUncheck {
				action = "marked incomplete"
			}
			fmt.Printf("✓ Step %d %s\n", stepID, action)

			// Show the Basecamp URL for easy access
			url := fmt.Sprintf("https://3.basecamp.com/%s/buckets/%s/card_tables/cards/%d/steps/%d",
				accountID, project, cardID, stepID)
			fmt.Printf("View at: %s\n", url)

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")
	cmd.Flags().StringVar(&noteOrReason, flagName, "", flagUsage)

	return cmd
}

// newStepCheckCmd creates the step check command
func newStepCheckCmd(f *factory.Factory) *cobra.Command {
	return createStepCommand(f, stepOperationCheck)
}

// newStepUncheckCmd creates the step uncheck command
func newStepUncheckCmd(f *factory.Factory) *cobra.Command {
	return createStepCommand(f, stepOperationUncheck)
}
