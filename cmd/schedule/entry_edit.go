package schedule

import (
	"fmt"
	"strconv"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/needmore/bc4/internal/utils"
	"github.com/spf13/cobra"
)

type entryEditOptions struct {
	title        string
	description  string
	startsAt     string
	endsAt       string
	allDay       *bool
	participants []string
	clearParticipants bool
}

func newEntryEditCmd(f *factory.Factory) *cobra.Command {
	opts := &entryEditOptions{}
	var allDayFlag bool
	var noAllDayFlag bool

	cmd := &cobra.Command{
		Use:   "edit <entry-id|URL>",
		Short: "Edit a schedule entry (calendar event)",
		Long: `Edit an existing calendar event in the project's schedule.

Only the specified fields will be updated; omitted fields remain unchanged.

You can specify the entry using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL`,
		Example: `  # Update event title
  bc4 schedule entry edit 12345 --title "New Meeting Title"

  # Update event times
  bc4 schedule entry edit 12345 --starts-at "2025-01-21T14:00:00" --ends-at "2025-01-21T15:00:00"

  # Change to all-day event
  bc4 schedule entry edit 12345 --all-day

  # Change to timed event
  bc4 schedule entry edit 12345 --no-all-day --starts-at "2025-01-21T14:00:00"

  # Update description
  bc4 schedule entry edit 12345 --description "Updated meeting notes"

  # Update participants
  bc4 schedule entry edit 12345 --participant "john@example.com" --participant "jane@example.com"

  # Clear all participants
  bc4 schedule entry edit 12345 --clear-participants`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Handle all-day flags
			if cmd.Flags().Changed("all-day") && allDayFlag {
				opts.allDay = &allDayFlag
			}
			if cmd.Flags().Changed("no-all-day") && noAllDayFlag {
				allDayFalse := false
				opts.allDay = &allDayFalse
			}
			return runEntryEdit(f, opts, args)
		},
	}

	cmd.Flags().StringVarP(&opts.title, "title", "t", "", "Update event title")
	cmd.Flags().StringVarP(&opts.description, "description", "d", "", "Update event description")
	cmd.Flags().StringVar(&opts.startsAt, "starts-at", "", "Update start date/time (YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS)")
	cmd.Flags().StringVar(&opts.endsAt, "ends-at", "", "Update end date/time (YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS)")
	cmd.Flags().BoolVar(&allDayFlag, "all-day", false, "Change to all-day event")
	cmd.Flags().BoolVar(&noAllDayFlag, "no-all-day", false, "Change to timed event")
	cmd.Flags().StringSliceVar(&opts.participants, "participant", nil, "Set participant (email or name, can be used multiple times)")
	cmd.Flags().BoolVar(&opts.clearParticipants, "clear-participants", false, "Remove all participants")

	return cmd
}

func runEntryEdit(f *factory.Factory, opts *entryEditOptions, args []string) error {
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

	// Build the update request
	req := api.ScheduleEntryUpdateRequest{}
	hasChanges := false

	if opts.title != "" {
		req.Summary = &opts.title
		hasChanges = true
	}

	if opts.description != "" {
		req.Description = &opts.description
		hasChanges = true
	}

	// Determine if this is going to be an all-day event for date parsing
	isAllDay := opts.allDay != nil && *opts.allDay

	if opts.startsAt != "" {
		startsAt, err := parseDateTime(opts.startsAt, isAllDay)
		if err != nil {
			return fmt.Errorf("invalid start date/time: %w", err)
		}
		req.StartsAt = &startsAt
		hasChanges = true
	}

	if opts.endsAt != "" {
		endsAt, err := parseDateTime(opts.endsAt, isAllDay)
		if err != nil {
			return fmt.Errorf("invalid end date/time: %w", err)
		}
		req.EndsAt = &endsAt
		hasChanges = true
	}

	if opts.allDay != nil {
		req.AllDay = opts.allDay
		hasChanges = true
	}

	// Handle participants
	if opts.clearParticipants {
		req.ParticipantIDs = []int64{} // Empty array to clear
		hasChanges = true
	} else if len(opts.participants) > 0 {
		// Create user resolver
		userResolver := utils.NewUserResolver(client.Client, projectID)

		// Resolve user identifiers to person IDs
		personIDs, err := userResolver.ResolveUsers(f.Context(), opts.participants)
		if err != nil {
			return fmt.Errorf("failed to resolve participants: %w", err)
		}

		req.ParticipantIDs = personIDs
		hasChanges = true
	}

	if !hasChanges {
		return fmt.Errorf("no changes specified. Use --help to see available options")
	}

	// Update the entry
	entry, err := scheduleOps.UpdateScheduleEntry(f.Context(), projectID, entryID, req)
	if err != nil {
		return fmt.Errorf("failed to update schedule entry: %w", err)
	}

	// Output confirmation
	fmt.Printf("#%d updated\n", entry.ID)

	return nil
}
