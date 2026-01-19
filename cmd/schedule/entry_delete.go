package schedule

import (
	"fmt"
	"strconv"

	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
)

type entryDeleteOptions struct {
	force bool
}

func newEntryDeleteCmd(f *factory.Factory) *cobra.Command {
	opts := &entryDeleteOptions{}

	cmd := &cobra.Command{
		Use:   "delete <entry-id|URL>",
		Short: "Delete a schedule entry (calendar event)",
		Long: `Delete (trash) a calendar event from the project's schedule.

The event will be moved to trash and can be recovered from the Basecamp web interface.

You can specify the entry using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL`,
		Example: `  # Delete an entry by ID
  bc4 schedule entry delete 12345

  # Delete without confirmation (if prompts are added in future)
  bc4 schedule entry delete 12345 --force`,
		Aliases: []string{"rm", "remove"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEntryDelete(f, opts, args)
		},
	}

	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}

func runEntryDelete(f *factory.Factory, opts *entryDeleteOptions, args []string) error {
	// Parse entry ID (could be numeric ID or URL)
	entryID, parsedURL, err := parser.ParseArgument(args[0])
	if err != nil {
		return fmt.Errorf("invalid entry ID or URL: %s", args[0])
	}

	// Apply URL overrides if provided
	if parsedURL != nil {
		if parsedURL.AccountID > 0 {
			f = f.WithAccount(strconv.FormatInt(parsedURL.AccountID, 10))
		}
		if parsedURL.ProjectID > 0 {
			f = f.WithProject(strconv.FormatInt(parsedURL.ProjectID, 10))
		}
	}

	// Get API client from factory
	client, err := f.ApiClient()
	if err != nil {
		return err
	}
	scheduleOps := client.Schedules()

	// Get resolved project ID
	projectID, err := f.ProjectID()
	if err != nil {
		return err
	}

	// Delete the entry
	err = scheduleOps.DeleteScheduleEntry(f.Context(), projectID, entryID)
	if err != nil {
		return fmt.Errorf("failed to delete schedule entry: %w", err)
	}

	// Output confirmation
	fmt.Printf("#%d deleted\n", entryID)

	return nil
}
