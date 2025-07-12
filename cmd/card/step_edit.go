package card

import (
	"github.com/spf13/cobra"
)

// newStepEditCmd creates the step edit command
func newStepEditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit CARD_ID STEP_ID",
		Short: "Edit a step's content",
		Long: `Edit the content of an existing step (subtask).

Examples:
  bc4 card step edit 123 456 --content "Updated step description"
  bc4 card step edit 123 456 --interactive`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement step edit functionality
			// 1. Parse card ID from args[0]
			// 2. Parse step ID from args[1]
			// 3. Get new content from flags or interactive editor
			// 4. Call API to update step
			// 5. Display success message
			return nil
		},
	}

	// TODO: Add flags for content and interactive mode
	cmd.Flags().String("content", "", "New content for the step")
	cmd.Flags().Bool("interactive", false, "Open interactive editor")

	return cmd
}
