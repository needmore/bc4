package card

import (
	"github.com/spf13/cobra"
)

// newStepCheckCmd creates the step check command
func newStepCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check CARD_ID STEP_ID",
		Short: "Mark a step as completed",
		Long: `Mark a step (subtask) as completed.

Examples:
  bc4 card step check 123 456
  bc4 card step check 123 456 --note "Fixed in PR #789"`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement step check functionality
			// 1. Parse card ID from args[0]
			// 2. Parse step ID from args[1]
			// 3. Get optional completion note from flags
			// 4. Call API to mark step as completed
			// 5. Display success message
			return nil
		},
	}

	// TODO: Add flags for completion notes
	cmd.Flags().String("note", "", "Add a completion note")

	return cmd
}
