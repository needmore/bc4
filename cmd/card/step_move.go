package card

import (
	"github.com/spf13/cobra"
)

// newStepMoveCmd creates the step move command
func newStepMoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move CARD_ID STEP_ID",
		Short: "Move a step to a different position",
		Long: `Move a step to a different position within the card.

Examples:
  bc4 card step move 123 456 --after 789
  bc4 card step move 123 456 --before 789
  bc4 card step move 123 456 --to-top
  bc4 card step move 123 456 --to-bottom`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement step move functionality
			// 1. Parse card ID from args[0]
			// 2. Parse step ID from args[1]
			// 3. Get position from flags
			// 4. Call API to reorder step
			// 5. Display success message
			return nil
		},
	}

	// TODO: Add flags for positioning
	cmd.Flags().String("after", "", "Move step after another step ID")
	cmd.Flags().String("before", "", "Move step before another step ID")
	cmd.Flags().Bool("to-top", false, "Move step to the top")
	cmd.Flags().Bool("to-bottom", false, "Move step to the bottom")

	return cmd
}
