package card

import (
	"github.com/spf13/cobra"
)

// newStepCmd creates the step management command
func newStepCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "step",
		Short: "Manage steps within cards",
		Long:  `Manage steps (subtasks) within cards, including adding, checking, and editing steps.`,
	}

	// Add subcommands
	cmd.AddCommand(newStepAddCmd())
	cmd.AddCommand(newStepListCmd())
	cmd.AddCommand(newStepCheckCmd())
	cmd.AddCommand(newStepUncheckCmd())
	cmd.AddCommand(newStepEditCmd())
	cmd.AddCommand(newStepMoveCmd())
	cmd.AddCommand(newStepAssignCmd())
	cmd.AddCommand(newStepDeleteCmd())

	return cmd
}
