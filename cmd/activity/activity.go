package activity

import (
	"github.com/needmore/bc4/internal/cmdutil"
	"github.com/needmore/bc4/internal/factory"
	"github.com/spf13/cobra"
)

// NewActivityCmd creates a new activity command
func NewActivityCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "activity",
		Short:   "View recent project activity",
		Aliases: []string{"events", "act"},
		Long: `View recent activity and events in a Basecamp project.

Activity shows recent changes to todos, messages, documents, and other
recordings in your project. Use this to track what's happening across
your team and stay up to date on project progress.`,
		Example: `  bc4 activity                        # List recent activity
  bc4 activity list                   # List recent activity (explicit)
  bc4 activity list --since "24h"     # Activity in last 24 hours
  bc4 activity list --type todo       # Only todo activity
  bc4 activity list --person "john"   # Activity by person
  bc4 activity list --format json     # Output as JSON
  bc4 activity watch                  # Watch for real-time activity
  bc4 activity watch --interval 10    # Poll every 10 seconds`,
	}

	// Enable suggestions for subcommand typos
	cmdutil.EnableSuggestions(cmd)

	// Add subcommands
	cmd.AddCommand(newListCmd(f))
	cmd.AddCommand(newWatchCmd(f))

	return cmd
}
