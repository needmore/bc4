package card

import (
	"github.com/spf13/cobra"
)

// NewCardCmd creates the card command
func NewCardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "card",
		Short: "Manage card tables and cards",
		Long: `Work with Basecamp card tables (kanban boards) and cards.

Card tables are Basecamp's take on kanban, perfect for managing reactive work
like software bugs, design requests, or other workflow-oriented tasks.`,
		Example: `  # List all card tables in the current project
  bc4 card list

  # View cards in a specific table
  bc4 card table "Bug Tracker"

  # Create a new card
  bc4 card add "Fix login issue"

  # View card details with steps
  bc4 card view 123

  # Move a card to a different column
  bc4 card move 123 --column "In Progress"`,
	}

	// Add subcommands
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newTableCmd())
	cmd.AddCommand(newViewCmd())
	cmd.AddCommand(newSetCmd())
	cmd.AddCommand(newAddCmd())
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newEditCmd())
	cmd.AddCommand(newMoveCmd())
	cmd.AddCommand(newAssignCmd())
	cmd.AddCommand(newUnassignCmd())
	cmd.AddCommand(newArchiveCmd())

	// Column management subcommands
	cmd.AddCommand(newColumnCmd())

	// Step management subcommands
	cmd.AddCommand(newStepCmd())

	return cmd
}
