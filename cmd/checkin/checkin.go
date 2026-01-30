package checkin

import (
	"github.com/needmore/bc4/internal/cmdutil"
	"github.com/needmore/bc4/internal/factory"
	"github.com/spf13/cobra"
)

// NewCheckinCmd creates the checkin command
func NewCheckinCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checkin",
		Short: "Manage automated check-ins (questionnaires)",
		Long: `Work with Basecamp automated check-ins.

Automated check-ins are recurring questions that Basecamp sends to your team
on a schedule. Team members can respond to check-ins, and everyone can see
the collected answers.

Examples:
  bc4 checkin list                     # List check-in questions in project
  bc4 checkin view 123                 # View a specific check-in question
  bc4 checkin answers 123              # View answers for a question
  bc4 checkin answer 123 "My update"   # Post an answer to a question
  bc4 checkin reminders                # List your pending check-in reminders
  bc4 checkin create "What did you work on today?" --schedule every_day
  bc4 checkin pause 123                # Pause a check-in question
  bc4 checkin resume 123               # Resume a paused question`,
		Aliases: []string{"checkins", "question", "questions"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Enable suggestions for subcommand typos
	cmdutil.EnableSuggestions(cmd)

	// Add subcommands
	cmd.AddCommand(newListCmd(f))
	cmd.AddCommand(newViewCmd(f))
	cmd.AddCommand(newRemindersCmd(f))
	cmd.AddCommand(newAnswersCmd(f))
	cmd.AddCommand(newAnswerCmd(f))
	cmd.AddCommand(newPauseCmd(f))
	cmd.AddCommand(newResumeCmd(f))
	cmd.AddCommand(newCreateCmd(f))
	cmd.AddCommand(newEditCmd(f))
	cmd.AddCommand(newNotifyCmd(f))

	return cmd
}
