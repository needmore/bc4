package schedule

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type viewOptions struct {
	jsonOutput bool
}

func newViewCmd(f *factory.Factory) *cobra.Command {
	opts := &viewOptions{}

	cmd := &cobra.Command{
		Use:   "view [schedule-id|URL]",
		Short: "View a specific schedule",
		Long: `View details of a specific schedule (calendar).

If no schedule ID is provided, shows the project's default schedule.

You can specify the schedule using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL`,
		Example: `  # View the project's schedule
  bc4 schedule view

  # View a specific schedule by ID
  bc4 schedule view 12345

  # Output as JSON
  bc4 schedule view --json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.jsonOutput = viper.GetBool("json")
			return runView(f, opts, args)
		},
	}

	return cmd
}

func runView(f *factory.Factory, opts *viewOptions, args []string) error {
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

	var scheduleID int64

	if len(args) > 0 {
		// Parse schedule ID (could be numeric ID or URL)
		parsedID, parsedURL, err := parser.ParseArgument(args[0])
		if err != nil {
			return fmt.Errorf("invalid schedule ID or URL: %s", args[0])
		}

		if parsedURL != nil {
			// Extract IDs from URL if provided
			if parsedURL.AccountID > 0 {
				f = f.WithAccount(strconv.FormatInt(parsedURL.AccountID, 10))
			}
			if parsedURL.ProjectID > 0 {
				f = f.WithProject(strconv.FormatInt(parsedURL.ProjectID, 10))
				projectID = strconv.FormatInt(parsedURL.ProjectID, 10)
			}
		}
		scheduleID = parsedID
	} else {
		// Get the project's default schedule
		schedule, err := scheduleOps.GetProjectSchedule(f.Context(), projectID)
		if err != nil {
			return fmt.Errorf("failed to get project schedule: %w", err)
		}
		scheduleID = schedule.ID
	}

	// Get the full schedule details
	schedule, err := scheduleOps.GetSchedule(f.Context(), projectID, scheduleID)
	if err != nil {
		return fmt.Errorf("failed to get schedule: %w", err)
	}

	// Output
	if opts.jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(schedule)
	}

	// Human-readable output
	fmt.Printf("Schedule: %s\n", schedule.Title)
	fmt.Printf("ID: %d\n", schedule.ID)
	fmt.Printf("Entries: %d\n", schedule.EntriesCount)
	if schedule.Status != "" {
		fmt.Printf("Status: %s\n", schedule.Status)
	}
	if schedule.AppURL != "" {
		fmt.Printf("URL: %s\n", schedule.AppURL)
	}

	return nil
}
