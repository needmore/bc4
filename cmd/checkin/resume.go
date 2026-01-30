package checkin

import (
	"fmt"
	"strconv"

	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
)

func newResumeCmd(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume <question-id|URL>",
		Short: "Resume a paused check-in question",
		Long: `Resume a paused check-in question to start sending reminders again.

This reverses the effect of the pause command.`,
		Example: `  # Resume a paused check-in question
  bc4 checkin resume 12345

  # Resume using URL
  bc4 checkin resume "https://3.basecamp.com/123/buckets/456/questions/789"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runResume(f, args)
		},
	}

	return cmd
}

func runResume(f *factory.Factory, args []string) error {
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

	// Resume the question
	if err := questionOps.ResumeQuestion(f.Context(), projectID, parsedID); err != nil {
		return fmt.Errorf("failed to resume question: %w", err)
	}

	fmt.Printf("Check-in question %d has been resumed.\n", parsedID)
	return nil
}
