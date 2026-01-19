package schedule

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type entryViewOptions struct {
	jsonOutput bool
}

func newEntryViewCmd(f *factory.Factory) *cobra.Command {
	opts := &entryViewOptions{}

	cmd := &cobra.Command{
		Use:   "view <entry-id|URL>",
		Short: "View a schedule entry's details",
		Long: `View detailed information about a specific schedule entry (calendar event).

You can specify the entry using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL`,
		Example: `  # View entry by ID
  bc4 schedule entry view 12345

  # View entry from URL
  bc4 schedule entry view https://3.basecamp.com/.../schedule_entries/12345

  # Output as JSON
  bc4 schedule entry view 12345 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.jsonOutput = viper.GetBool("json")
			return runEntryView(f, opts, args)
		},
	}

	return cmd
}

func runEntryView(f *factory.Factory, opts *entryViewOptions, args []string) error {
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

	// Get the entry details
	entry, err := scheduleOps.GetScheduleEntry(f.Context(), projectID, entryID)
	if err != nil {
		return fmt.Errorf("failed to get schedule entry: %w", err)
	}

	// Output
	if opts.jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(entry)
	}

	// Human-readable output
	title := entry.Title
	if title == "" {
		title = entry.Summary
	}
	fmt.Printf("Event: %s\n", title)
	fmt.Printf("ID: %d\n", entry.ID)

	// Format and display date/time
	if entry.AllDay {
		fmt.Println("Type: All-day event")
	}

	startTime, err := time.Parse(time.RFC3339, entry.StartsAt)
	if err == nil {
		if entry.AllDay {
			fmt.Printf("Date: %s\n", startTime.Format("Monday, January 2, 2006"))
		} else {
			fmt.Printf("Starts: %s\n", startTime.Format("Monday, January 2, 2006 at 3:04 PM"))
		}
	}

	if entry.EndsAt != "" {
		endTime, err := time.Parse(time.RFC3339, entry.EndsAt)
		if err == nil {
			if !entry.AllDay {
				fmt.Printf("Ends: %s\n", endTime.Format("Monday, January 2, 2006 at 3:04 PM"))
			} else if startTime.Format("2006-01-02") != endTime.Format("2006-01-02") {
				// Multi-day all-day event
				fmt.Printf("Through: %s\n", endTime.Format("Monday, January 2, 2006"))
			}
		}
	}

	// Description
	if entry.Description != "" {
		fmt.Printf("\nDescription:\n%s\n", entry.Description)
	}

	// Participants
	if len(entry.Participants) > 0 {
		fmt.Printf("\nParticipants (%d):\n", len(entry.Participants))
		for _, p := range entry.Participants {
			fmt.Printf("  - %s", p.Name)
			if p.EmailAddress != "" {
				fmt.Printf(" <%s>", p.EmailAddress)
			}
			fmt.Println()
		}
	}

	// Recurrence info
	if entry.Recurrence != nil {
		fmt.Printf("\nRecurrence: %s\n", formatRecurrence(entry.Recurrence))
	}

	// Status
	if entry.Status != "" && entry.Status != "active" {
		fmt.Printf("\nStatus: %s\n", entry.Status)
	}

	// URL
	if entry.AppURL != "" {
		fmt.Printf("\nURL: %s\n", entry.AppURL)
	}

	// Comments
	if entry.CommentsCount > 0 {
		fmt.Printf("\nComments: %d\n", entry.CommentsCount)
	}

	return nil
}

// formatRecurrence formats a recurrence schedule for display
func formatRecurrence(r *api.Recurrence) string {
	if r == nil {
		return ""
	}

	var result string
	switch r.Frequency {
	case "every_day":
		result = "Daily"
	case "every_week":
		result = "Weekly"
		if len(r.Days) > 0 {
			days := []string{}
			dayNames := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
			for _, d := range r.Days {
				if d >= 0 && d < 7 {
					days = append(days, dayNames[d])
				}
			}
			if len(days) > 0 {
				result += " on " + strings.Join(days, ", ")
			}
		}
	case "every_month":
		result = "Monthly"
		if r.Week != "" {
			result += " on the " + r.Week + " week"
		}
	case "every_year":
		result = "Yearly"
	default:
		result = r.Frequency
	}

	if r.EndDate != "" {
		result += " until " + r.EndDate
	}

	return result
}
