package schedule

import (
	"github.com/needmore/bc4/internal/cmdutil"
	"github.com/needmore/bc4/internal/factory"
	"github.com/spf13/cobra"
)

// NewScheduleCmd creates the schedule command
func NewScheduleCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schedule",
		Short: "Work with Basecamp schedules - calendar events and entries",
		Long: `Work with Basecamp schedules and calendar events.

Each Basecamp project can have a schedule (calendar) containing events.
Events can be single or recurring, all-day or timed, and can have participants.

Examples:
  bc4 schedule list                      # List schedules in current project
  bc4 schedule view 123                  # View schedule details
  bc4 schedule entry list                # List calendar events
  bc4 schedule entry list --upcoming     # List upcoming events
  bc4 schedule entry view 456            # View event details
  bc4 schedule entry create "Meeting"    # Create a new event`,
		Aliases: []string{"schedules", "cal", "calendar"},
	}

	// Enable suggestions for subcommand typos
	cmdutil.EnableSuggestions(cmd)

	// Add subcommands
	cmd.AddCommand(newListCmd(f))
	cmd.AddCommand(newViewCmd(f))
	cmd.AddCommand(newEntryCmd(f))

	return cmd
}
