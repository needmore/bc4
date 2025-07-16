package card

import (
	"fmt"
	"strconv"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
)

// newStepDeleteCmd creates the step delete command
func newStepDeleteCmd(f *factory.Factory) *cobra.Command {
	var accountID string
	var projectID string

	cmd := &cobra.Command{
		Use:   "delete [CARD_ID or URL] STEP_ID",
		Short: "Delete a step from a card",
		Long: `Delete a step (subtask) from a card.

You can specify the card using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345")

Examples:
  bc4 card step delete 123 456
  bc4 card step delete 123 456 --force
  bc4 card step delete https://3.basecamp.com/1234567/buckets/89012345/card_tables/cards/12345 456`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse card ID (could be numeric ID or URL)
			cardID, parsedURL, err := parser.ParseArgument(args[0])
			if err != nil {
				return fmt.Errorf("invalid card ID or URL: %s", args[0])
			}

			// Parse step ID
			stepID, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid step ID: %s", args[1])
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

			// Check if force flag is set
			force, _ := cmd.Flags().GetBool("force")

			// If not forcing, confirm deletion
			if !force {
				// Get the card to show step details
				card, err := client.Cards().GetCard(f.Context(), resolvedProjectID, cardID)
				if err != nil {
					return fmt.Errorf("failed to get card: %w", err)
				}

				// Find the step to confirm deletion
				var stepToDelete *api.Step
				for _, step := range card.Steps {
					if step.ID == stepID {
						stepToDelete = &step
						break
					}
				}

				if stepToDelete == nil {
					return fmt.Errorf("step with ID %d not found in card", stepID)
				}

				// Ask for confirmation
				fmt.Printf("Delete step #%d: \"%s\"? (y/N) ", stepID, stepToDelete.Title)
				var response string
				_, _ = fmt.Scanln(&response)
				if response != "y" && response != "Y" {
					fmt.Println("Deletion cancelled")
					return nil
				}
			}

			// Delete the step
			err = client.Steps().DeleteStep(f.Context(), resolvedProjectID, stepID)
			if err != nil {
				return fmt.Errorf("failed to delete step: %w", err)
			}

			fmt.Printf("Step #%d deleted\n", stepID)
			return nil
		},
	}

	// TODO: Add flags for force delete
	cmd.Flags().Bool("force", false, "Delete without confirmation")
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")

	return cmd
}
