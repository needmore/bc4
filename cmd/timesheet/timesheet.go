package timesheet

import (
	"github.com/needmore/bc4/internal/cmdutil"
	"github.com/needmore/bc4/internal/factory"
	"github.com/spf13/cobra"
)

// NewTimesheetCmd creates a new timesheet command
func NewTimesheetCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "timesheet",
		Short:   "View timesheet entries",
		Aliases: []string{"time", "ts"},
		Long: `View and manage timesheet entries in Basecamp projects.

Timesheet entries track time spent on work items. This command provides
read-only access to view entries across projects or for specific recordings.

Note: The Basecamp API currently only supports reading timesheet entries.
Creating or updating entries must be done through the Basecamp web interface.`,
		Example: `  bc4 timesheet list                        # List entries for current project
  bc4 timesheet list --project 123456       # List entries for specific project
  bc4 timesheet list --since 7d             # Entries from last 7 days
  bc4 timesheet list --person "john"        # Filter by person
  bc4 timesheet report                      # Account-wide report
  bc4 timesheet report --start 2024-01-01   # Report with date range
  bc4 timesheet report --end 2024-01-31     # Report ending on date`,
	}

	// Enable suggestions for subcommand typos
	cmdutil.EnableSuggestions(cmd)

	// Add subcommands
	cmd.AddCommand(newListCmd(f))
	cmd.AddCommand(newReportCmd(f))

	return cmd
}
