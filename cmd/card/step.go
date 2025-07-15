package card

import (
	"github.com/needmore/bc4/internal/factory"
	"github.com/spf13/cobra"
)

// newStepCmd creates the step management command
func newStepCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "step",
		Short: "Manage steps within cards",
		Long:  `Manage steps (subtasks) within cards, including adding, checking, and editing steps.`,
	}

	// Add subcommands
	cmd.AddCommand(newStepAddCmd(f))
	cmd.AddCommand(newStepListCmd(f))
	cmd.AddCommand(newStepCheckCmd(f))
	cmd.AddCommand(newStepUncheckCmd(f))
	cmd.AddCommand(newStepEditCmd(f))
	cmd.AddCommand(newStepMoveCmd(f))
	cmd.AddCommand(newStepAssignCmd(f))
	cmd.AddCommand(newStepDeleteCmd(f))

	return cmd
}
