package card

import (
	"github.com/spf13/cobra"
)

// newStepUncheckCmd creates the step uncheck command
func newStepUncheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uncheck CARD_ID STEP_ID",
		Short: "Mark a step as incomplete",
		Long: `Mark a completed step (subtask) as incomplete again.

Examples:
  bc4 card step uncheck 123 456
  bc4 card step uncheck 123 456 --reason "Needs rework"`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement step uncheck functionality
			// 1. Parse card ID from args[0]
			// 2. Parse step ID from args[1]
			// 3. Get optional reason from flags
			// 4. Call API to mark step as incomplete
			// 5. Display success message
			return nil
		},
	}

	// TODO: Add flags for reason
	cmd.Flags().String("reason", "", "Add a reason for marking incomplete")

	return cmd
}
