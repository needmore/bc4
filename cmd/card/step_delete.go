package card

import (
	"github.com/spf13/cobra"
)

// newStepDeleteCmd creates the step delete command
func newStepDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete CARD_ID STEP_ID",
		Short: "Delete a step from a card",
		Long: `Delete a step (subtask) from a card.

Examples:
  bc4 card step delete 123 456
  bc4 card step delete 123 456 --force`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement step delete functionality
			// 1. Parse card ID from args[0]
			// 2. Parse step ID from args[1]
			// 3. Check for force flag or prompt for confirmation
			// 4. Call API to delete step
			// 5. Display success message
			return nil
		},
	}

	// TODO: Add flags for force delete
	cmd.Flags().Bool("force", false, "Delete without confirmation")

	return cmd
}
