package card

import (
	"github.com/spf13/cobra"
)

// newStepAssignCmd creates the step assign command
func newStepAssignCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assign CARD_ID STEP_ID USER",
		Short: "Assign a step to a user",
		Long: `Assign a step (subtask) to a specific user.

Examples:
  bc4 card step assign 123 456 @jane
  bc4 card step assign 123 456 "Jane Doe"
  bc4 card step assign 123 456 jane@example.com`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement step assign functionality
			// 1. Parse card ID from args[0]
			// 2. Parse step ID from args[1]
			// 3. Parse user identifier from args[2]
			// 4. Resolve user (by handle, name, or email)
			// 5. Call API to assign step
			// 6. Display success message
			return nil
		},
	}

	// TODO: Add flags for unassigning
	cmd.Flags().Bool("unassign", false, "Remove current assignee instead")

	return cmd
}
