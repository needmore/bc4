package schedule

import (
	"fmt"
	"strings"
	"time"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/utils"
	"github.com/spf13/cobra"
)

type entryCreateOptions struct {
	description  string
	startsAt     string
	endsAt       string
	allDay       bool
	participants []string
	notify       bool
}

func newEntryCreateCmd(f *factory.Factory) *cobra.Command {
	opts := &entryCreateOptions{}

	cmd := &cobra.Command{
		Use:   "create <title>",
		Short: "Create a new schedule entry (calendar event)",
		Long: `Create a new calendar event in the project's schedule.

Events can be:
- Timed events with specific start and end times
- All-day events using --all-day flag
- With participants from the project

Date/time formats:
- Date only: 2025-01-15 (assumes all-day or start of day)
- With time: 2025-01-15T14:30:00
- Relative: today, tomorrow, next-monday`,
		Example: `  # Create a simple event (prompts for time)
  bc4 schedule entry create "Team Meeting"

  # Create an event with specific times
  bc4 schedule entry create "Sprint Review" --starts-at "2025-01-20T14:00:00" --ends-at "2025-01-20T15:00:00"

  # Create an all-day event
  bc4 schedule entry create "Company Holiday" --starts-at "2025-01-20" --all-day

  # Create an event with participants
  bc4 schedule entry create "1:1 Meeting" --starts-at "2025-01-20T10:00:00" --ends-at "2025-01-20T10:30:00" --participant "john@example.com"

  # Create a multi-day event
  bc4 schedule entry create "Conference" --starts-at "2025-01-20" --ends-at "2025-01-22" --all-day

  # Create and notify participants
  bc4 schedule entry create "Urgent Meeting" --starts-at "2025-01-20T09:00:00" --notify`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEntryCreate(f, opts, args)
		},
	}

	cmd.Flags().StringVarP(&opts.description, "description", "d", "", "Event description")
	cmd.Flags().StringVar(&opts.startsAt, "starts-at", "", "Start date/time (YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS)")
	cmd.Flags().StringVar(&opts.endsAt, "ends-at", "", "End date/time (YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS)")
	cmd.Flags().BoolVar(&opts.allDay, "all-day", false, "Create an all-day event")
	cmd.Flags().StringSliceVar(&opts.participants, "participant", nil, "Add participant (email or name, can be used multiple times)")
	cmd.Flags().BoolVar(&opts.notify, "notify", false, "Notify participants about the event")

	_ = cmd.MarkFlagRequired("starts-at")

	return cmd
}

func runEntryCreate(f *factory.Factory, opts *entryCreateOptions, args []string) error {
	title := args[0]

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

	// Parse and validate dates
	startsAt, err := parseDateTime(opts.startsAt, opts.allDay)
	if err != nil {
		return fmt.Errorf("invalid start date/time: %w", err)
	}

	var endsAt string
	if opts.endsAt != "" {
		endsAt, err = parseDateTime(opts.endsAt, opts.allDay)
		if err != nil {
			return fmt.Errorf("invalid end date/time: %w", err)
		}
	} else {
		// Default end time
		if opts.allDay {
			// For all-day events, end is same as start (single day)
			endsAt = startsAt
		} else {
			// Default to 1 hour duration
			startTime, _ := time.Parse(time.RFC3339, startsAt)
			endsAt = startTime.Add(time.Hour).Format(time.RFC3339)
		}
	}

	// Build the request
	req := api.ScheduleEntryCreateRequest{
		Summary:     title,
		Description: opts.description,
		StartsAt:    startsAt,
		EndsAt:      endsAt,
		AllDay:      opts.allDay,
		Notify:      opts.notify,
	}

	// Handle participants
	if len(opts.participants) > 0 {
		// Create user resolver
		userResolver := utils.NewUserResolver(client.Client, projectID)

		// Resolve user identifiers to person IDs
		personIDs, err := userResolver.ResolveUsers(f.Context(), opts.participants)
		if err != nil {
			return fmt.Errorf("failed to resolve participants: %w", err)
		}

		req.ParticipantIDs = personIDs
	}

	// TODO: Add support for recurring events via --recurrence flag
	// This would populate the recurrence_schedule field in the request
	// Example: --recurrence "every_week" --days 1,2,3,4,5

	// Create the entry
	entry, err := scheduleOps.CreateScheduleEntry(f.Context(), projectID, schedule.ID, req)
	if err != nil {
		return fmt.Errorf("failed to create schedule entry: %w", err)
	}

	// Output the entry ID (GitHub CLI style)
	fmt.Printf("#%d\n", entry.ID)

	return nil
}

// parseDateTime parses various date/time formats and returns RFC3339 format
func parseDateTime(input string, allDay bool) (string, error) {
	input = strings.TrimSpace(input)

	// Handle relative dates
	now := time.Now()
	switch strings.ToLower(input) {
	case "today":
		if allDay {
			return now.Format("2006-01-02"), nil
		}
		return time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location()).Format(time.RFC3339), nil
	case "tomorrow":
		tomorrow := now.AddDate(0, 0, 1)
		if allDay {
			return tomorrow.Format("2006-01-02"), nil
		}
		return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 9, 0, 0, 0, now.Location()).Format(time.RFC3339), nil
	case "next-monday":
		daysUntilMonday := (8 - int(now.Weekday())) % 7
		if daysUntilMonday == 0 {
			daysUntilMonday = 7
		}
		monday := now.AddDate(0, 0, daysUntilMonday)
		if allDay {
			return monday.Format("2006-01-02"), nil
		}
		return time.Date(monday.Year(), monday.Month(), monday.Day(), 9, 0, 0, 0, now.Location()).Format(time.RFC3339), nil
	}

	// Try RFC3339 format first
	if t, err := time.Parse(time.RFC3339, input); err == nil {
		if allDay {
			return t.Format("2006-01-02"), nil
		}
		return t.Format(time.RFC3339), nil
	}

	// Try ISO 8601 with seconds
	if t, err := time.Parse("2006-01-02T15:04:05", input); err == nil {
		if allDay {
			return t.Format("2006-01-02"), nil
		}
		return t.Format(time.RFC3339), nil
	}

	// Try ISO 8601 without seconds
	if t, err := time.Parse("2006-01-02T15:04", input); err == nil {
		if allDay {
			return t.Format("2006-01-02"), nil
		}
		return t.Format(time.RFC3339), nil
	}

	// Try date-only format
	if t, err := time.Parse("2006-01-02", input); err == nil {
		if allDay {
			return t.Format("2006-01-02"), nil
		}
		// Default to 9 AM for date-only input on timed events
		return time.Date(t.Year(), t.Month(), t.Day(), 9, 0, 0, 0, time.Local).Format(time.RFC3339), nil
	}

	return "", fmt.Errorf("unrecognized date/time format: %s (use YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS)", input)
}
