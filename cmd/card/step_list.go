package card

import (
	"github.com/spf13/cobra"
)

// newStepListCmd creates the step list command
func newStepListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list CARD_ID",
		Short: "List all steps in a card",
		Long: `List all steps (subtasks) in a card, showing their status and assignees.

Examples:
  bc4 card step list 123
  bc4 card step list 123 --completed
  bc4 card step list 123 --format json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement step list functionality
			// 1. Parse card ID from args[0]
			// 2. Get filter options from flags
			// 3. Call API to fetch card with steps
			// 4. Filter steps based on flags
			// 5. Display steps in table or requested format
			return nil
		},
	}

	// TODO: Add flags for filtering and formatting
	cmd.Flags().Bool("completed", false, "Show only completed steps")
	cmd.Flags().Bool("pending", false, "Show only pending steps")
	cmd.Flags().String("assignee", "", "Filter by assignee")
	cmd.Flags().String("format", "table", "Output format: table, json, csv")

	return cmd
}
