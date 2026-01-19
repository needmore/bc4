package schedule

import (
	"github.com/needmore/bc4/internal/cmdutil"
	"github.com/needmore/bc4/internal/factory"
	"github.com/spf13/cobra"
)

// newEntryCmd creates the entry management command
func newEntryCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "entry",
		Short: "Manage schedule entries (calendar events)",
		Long: `Manage schedule entries (calendar events), including listing, viewing, creating, editing, and deleting events.

Schedule entries can be:
- Single events or recurring (daily, weekly, monthly, yearly)
- All-day events or timed events with specific start/end times
- Assigned to participants from the project

Examples:
  bc4 schedule entry list                    # List all entries
  bc4 schedule entry list --upcoming         # List upcoming events
  bc4 schedule entry list --past             # List past events
  bc4 schedule entry view 123                # View entry details
  bc4 schedule entry create "Team Meeting"   # Create new event
  bc4 schedule entry edit 123                # Edit an entry
  bc4 schedule entry delete 123              # Delete an entry`,
		Aliases: []string{"entries", "event", "events"},
	}

	// Enable suggestions for subcommand typos
	cmdutil.EnableSuggestions(cmd)

	// Add subcommands
	cmd.AddCommand(newEntryListCmd(f))
	cmd.AddCommand(newEntryViewCmd(f))
	cmd.AddCommand(newEntryCreateCmd(f))
	cmd.AddCommand(newEntryEditCmd(f))
	cmd.AddCommand(newEntryDeleteCmd(f))

	return cmd
}
