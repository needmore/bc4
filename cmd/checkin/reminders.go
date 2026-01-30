package checkin

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/ui/tableprinter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type remindersOptions struct {
	jsonOutput bool
}

func newRemindersCmd(f *factory.Factory) *cobra.Command {
	opts := &remindersOptions{}

	cmd := &cobra.Command{
		Use:   "reminders",
		Short: "List your pending check-in reminders",
		Long: `List all pending check-in reminders for your account.

Shows check-ins that are due for you to answer.`,
		Example: `  # List pending reminders
  bc4 checkin reminders

  # Output as JSON
  bc4 checkin reminders --json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.jsonOutput = viper.GetBool("json")
			return runReminders(f, opts)
		},
	}

	return cmd
}

func runReminders(f *factory.Factory, opts *remindersOptions) error {
	client, err := f.ApiClient()
	if err != nil {
		return err
	}
	questionOps := client.Questions()

	// List all reminders
	reminders, err := questionOps.ListMyReminders(f.Context())
	if err != nil {
		return fmt.Errorf("failed to list reminders: %w", err)
	}

	if opts.jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(reminders)
	}

	if len(reminders) == 0 {
		fmt.Println("No pending check-in reminders.")
		return nil
	}

	// Create table
	table := tableprinter.New(os.Stdout)

	// Add headers
	if table.IsTTY() {
		table.AddHeader("QUESTION ID", "QUESTION", "PROJECT", "REMIND AT", "GROUP ON")
	} else {
		table.AddHeader("QUESTION_ID", "QUESTION", "PROJECT", "REMIND_AT", "GROUP_ON")
	}

	now := time.Now()

	for _, r := range reminders {
		// Question ID
		table.AddIDField(strconv.FormatInt(r.QuestionID, 10), "active")

		// Question title
		questionTitle := ""
		if r.Question != nil {
			questionTitle = r.Question.Title
			if len(questionTitle) > 40 {
				questionTitle = questionTitle[:37] + "..."
			}
		}
		table.AddField(questionTitle)

		// Project name
		projectName := ""
		if r.Bucket != nil {
			projectName = r.Bucket.Name
		}
		table.AddField(projectName)

		// Remind at
		table.AddTimeField(now, r.RemindAt)

		// Group on (date)
		table.AddField(r.GroupOn)

		table.EndRow()
	}

	return table.Render()
}
