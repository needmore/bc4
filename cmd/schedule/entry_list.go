package schedule

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/ui/tableprinter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type entryListOptions struct {
	upcoming   bool
	past       bool
	startDate  string
	endDate    string
	jsonOutput bool
}

func newEntryListCmd(f *factory.Factory) *cobra.Command {
	opts := &entryListOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List schedule entries (calendar events)",
		Long: `List calendar events in the project's schedule.

By default, lists all entries. Use filters to narrow down results:
- --upcoming: Show only future events
- --past: Show only past events
- --range: Filter by date range (YYYY-MM-DD format)`,
		Example: `  # List all events
  bc4 schedule entry list

  # List upcoming events only
  bc4 schedule entry list --upcoming

  # List past events only
  bc4 schedule entry list --past

  # List events in a date range
  bc4 schedule entry list --range 2025-01-01 2025-01-31

  # Output as JSON
  bc4 schedule entry list --json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.jsonOutput = viper.GetBool("json")

			// Handle --range flag with two positional values
			rangeArgs, _ := cmd.Flags().GetStringSlice("range")
			if len(rangeArgs) == 2 {
				opts.startDate = rangeArgs[0]
				opts.endDate = rangeArgs[1]
			} else if len(rangeArgs) == 1 {
				return fmt.Errorf("--range requires both start and end dates (e.g., --range 2025-01-01,2025-01-31)")
			}

			return runEntryList(f, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.upcoming, "upcoming", false, "Show only upcoming events")
	cmd.Flags().BoolVar(&opts.past, "past", false, "Show only past events")
	cmd.Flags().StringSlice("range", nil, "Filter by date range: --range START_DATE,END_DATE (YYYY-MM-DD)")

	return cmd
}

func runEntryList(f *factory.Factory, opts *entryListOptions) error {
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

	// Get the project's schedule
	schedule, err := scheduleOps.GetProjectSchedule(f.Context(), projectID)
	if err != nil {
		return fmt.Errorf("failed to get schedule: %w", err)
	}

	// Fetch entries based on filters
	var entries []api.ScheduleEntry

	if opts.upcoming {
		entries, err = scheduleOps.GetUpcomingScheduleEntries(f.Context(), projectID, schedule.ID)
	} else if opts.past {
		entries, err = scheduleOps.GetPastScheduleEntries(f.Context(), projectID, schedule.ID)
	} else if opts.startDate != "" && opts.endDate != "" {
		entries, err = scheduleOps.GetScheduleEntriesInRange(f.Context(), projectID, schedule.ID, opts.startDate, opts.endDate)
	} else {
		entries, err = scheduleOps.GetScheduleEntries(f.Context(), projectID, schedule.ID)
	}

	if err != nil {
		return fmt.Errorf("failed to fetch schedule entries: %w", err)
	}

	if len(entries) == 0 {
		if opts.jsonOutput {
			fmt.Println("[]")
			return nil
		}
		fmt.Println("No schedule entries found.")
		return nil
	}

	// Output
	if opts.jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(entries)
	}

	// Table output
	table := tableprinter.New(os.Stdout)

	table.AddHeader("ID", "TITLE", "DATE", "TIME", "ALL-DAY", "PARTICIPANTS")

	for _, entry := range entries {
		table.AddField(fmt.Sprintf("%d", entry.ID))

		// Truncate long titles
		title := entry.Title
		if title == "" {
			title = entry.Summary
		}
		if len(title) > 40 {
			title = title[:37] + "..."
		}
		table.AddField(title)

		// Format date and time
		dateStr, timeStr := formatDateTime(entry.StartsAt, entry.EndsAt, entry.AllDay)
		table.AddField(dateStr)
		table.AddField(timeStr)

		// All-day indicator
		if entry.AllDay {
			table.AddField("Yes")
		} else {
			table.AddField("No")
		}

		// Participants count
		if len(entry.Participants) > 0 {
			table.AddField(fmt.Sprintf("%d", len(entry.Participants)))
		} else {
			table.AddField("-")
		}

		table.EndRow()
	}

	return table.Render()
}

// formatDateTime formats the start and end times for display
func formatDateTime(startsAt, endsAt string, allDay bool) (dateStr, timeStr string) {
	// Parse the timestamps
	startTime, err := time.Parse(time.RFC3339, startsAt)
	if err != nil {
		// Try alternative format
		startTime, err = time.Parse("2006-01-02", startsAt)
		if err != nil {
			return startsAt, ""
		}
	}

	endTime, _ := time.Parse(time.RFC3339, endsAt)

	// Format date
	dateStr = startTime.Format("Jan 02, 2006")

	// Check if spans multiple days
	if !endTime.IsZero() && startTime.Format("2006-01-02") != endTime.Format("2006-01-02") {
		dateStr = fmt.Sprintf("%s - %s", startTime.Format("Jan 02"), endTime.Format("Jan 02, 2006"))
	}

	// Format time
	if allDay {
		timeStr = "All day"
	} else {
		timeStr = startTime.Format("3:04 PM")
		if !endTime.IsZero() {
			timeStr = fmt.Sprintf("%s - %s", startTime.Format("3:04 PM"), endTime.Format("3:04 PM"))
		}
	}

	return dateStr, timeStr
}

// formatParticipants formats participant names for display
func formatParticipants(participants []api.Person) string {
	if len(participants) == 0 {
		return "-"
	}

	names := make([]string, len(participants))
	for i, p := range participants {
		names[i] = p.Name
	}

	result := strings.Join(names, ", ")
	if len(result) > 30 {
		return fmt.Sprintf("%s... (+%d)", names[0], len(participants)-1)
	}
	return result
}
