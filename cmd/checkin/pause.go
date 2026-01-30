package checkin

import (
	"fmt"
	"strconv"

	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
)

func newPauseCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pause <question-id|URL>",
		Short: "Pause a check-in question",
		Long: `Pause a check-in question to temporarily stop it from sending reminders.

A paused question will not send any reminders until it is resumed.`,
		Example: `  # Pause a check-in question
  bc4 checkin pause 12345

  # Pause using URL
  bc4 checkin pause "https://3.basecamp.com/123/buckets/456/questions/789"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPause(f, args)
		},
	}

	return cmd
}

func runPause(f *factory.Factory, args []string) error {
	client, err := f.ApiClient()
	if err != nil {
		return err
	}
	questionOps := client.Questions()

	projectID, err := f.ProjectID()
	if err != nil {
		return err
	}

	// Parse question ID (could be numeric ID or URL)
	parsedID, parsedURL, err := parser.ParseArgument(args[0])
	if err != nil {
		return fmt.Errorf("invalid question ID or URL: %s", args[0])
	}

	if parsedURL != nil {
		if parsedURL.AccountID > 0 {
			f = f.WithAccount(strconv.FormatInt(parsedURL.AccountID, 10))
		}
		if parsedURL.ProjectID > 0 {
			projectID = strconv.FormatInt(parsedURL.ProjectID, 10)
		}
	}

	// Pause the question
	if err := questionOps.PauseQuestion(f.Context(), projectID, parsedID); err != nil {
		return fmt.Errorf("failed to pause question: %w", err)
	}

	fmt.Printf("Check-in question %d has been paused.\n", parsedID)
	return nil
}
