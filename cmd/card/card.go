package card

import (
	"github.com/needmore/bc4/internal/cmdutil"
	"github.com/needmore/bc4/internal/factory"
	"github.com/spf13/cobra"
)

// NewCardCmd creates the card command
func NewCardCmd(f *factory.Factory) *cobra.Command {
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

	// Enable suggestions for subcommand typos
	cmdutil.EnableSuggestions(cmd)

	// Add subcommands
	cmd.AddCommand(newListCmd(f))
	cmd.AddCommand(newTableCmd(f))
	cmd.AddCommand(newViewCmd(f))
	cmd.AddCommand(newSetCmd(f))
	cmd.AddCommand(newAddCmd(f))
	cmd.AddCommand(newCreateCmd(f))
	cmd.AddCommand(newEditCmd(f))
	cmd.AddCommand(newMoveCmd(f))
	cmd.AddCommand(newAssignCmd(f))
	cmd.AddCommand(newUnassignCmd(f))
	cmd.AddCommand(newArchiveCmd(f))

	// Column management subcommands
	cmd.AddCommand(newColumnCmd(f))

	// Step management subcommands
	cmd.AddCommand(newStepCmd(f))

	// Attachments subcommand
	cmd.AddCommand(newAttachmentsCmd(f))
	cmd.AddCommand(newDownloadAttachmentsCmd(f))

	return cmd
}
