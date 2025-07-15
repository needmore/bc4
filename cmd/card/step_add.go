package card

import (
	"fmt"
	"strconv"

	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
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

			// Get API client from factory (for auth check)
			_, err = f.ApiClient()
			if err != nil {
				return err
			}

			// TODO: Implement step add functionality
			// Get step content from args[1]
			stepContent := args[1]
			fmt.Printf("Would add step \"%s\" to card ID %d\n", stepContent, cardID)
			return fmt.Errorf("step add not yet implemented")
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
