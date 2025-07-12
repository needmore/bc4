package card

import (
	"github.com/spf13/cobra"
)

// newStepAddCmd creates the step add command
func newStepAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add CARD_ID \"Step content\"",
		Short: "Add a new step to a card",
		Long: `Add a new step (subtask) to an existing card.

Examples:
  bc4 card step add 123 "Review PR comments"
  bc4 card step add 456 "Update documentation" --assign @jane`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement step add functionality
			// 1. Parse card ID from args[0]
			// 2. Get step content from args[1]
			// 3. Get optional assignee from flags
			// 4. Call API to create step
			// 5. Display success message
			return nil
		},
	}

	// TODO: Add flags for assignee, position, etc.
	cmd.Flags().String("assign", "", "Assign the step to a user")
	cmd.Flags().String("after", "", "Position step after another step ID")
	cmd.Flags().String("before", "", "Position step before another step ID")

	return cmd
}
