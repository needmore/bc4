package timesheet

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
	"github.com/needmore/bc4/internal/ui/tableprinter"
	"github.com/spf13/cobra"
)

func newListCmd(f *factory.Factory) *cobra.Command {
	var (
		accountID   string
		projectID   string
		personStr   string
		sinceStr    string
		formatStr   string
		recordingID int64
	)

	cmd := &cobra.Command{
		Use:     "list [project]",
		Short:   "List timesheet entries",
		Long:    `List timesheet entries for a project or recording.`,
		Aliases: []string{"ls"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse project argument if provided (could be URL or ID)
			if len(args) > 0 {
				if parser.IsBasecampURL(args[0]) {
					parsed, err := parser.ParseBasecampURL(args[0])
					if err != nil {
						return fmt.Errorf("invalid Basecamp URL: %w", err)
					}
					if parsed.ResourceType != parser.ResourceTypeProject {
						return fmt.Errorf("URL is not for a project: %s", args[0])
					}
					// Override factory with URL-provided values
					if parsed.AccountID > 0 {
						f = f.WithAccount(strconv.FormatInt(parsed.AccountID, 10))
					}
					if parsed.ProjectID > 0 {
						f = f.WithProject(strconv.FormatInt(parsed.ProjectID, 10))
					}
				} else {
					// Treat as project ID
					f = f.WithProject(args[0])
				}
			}

			// Apply flag overrides (flags take precedence over URL)
			if accountID != "" {
				f = f.WithAccount(accountID)
			}
			if projectID != "" {
				f = f.WithProject(projectID)
			}

			// Parse --since before fetching so invalid input fails early
			var sinceDate time.Time
			if sinceStr != "" {
				var err error
				sinceDate, err = parseSince(sinceStr)
				if err != nil {
					return fmt.Errorf("invalid --since value: %w", err)
				}
			}

			// Get API client from factory
			client, err := f.ApiClient()
			if err != nil {
				return err
			}

			// Get resolved project ID
			resolvedProjectID, err := f.ProjectID()
			if err != nil {
				return err
			}

			// Fetch timesheet entries
			var entries []api.TimesheetEntry
			if recordingID > 0 {
				// Get entries for a specific recording
				entries, err = client.Timesheets().GetRecordingTimesheet(cmd.Context(), resolvedProjectID, recordingID)
			} else {
				// Get entries for the project
				entries, err = client.Timesheets().GetProjectTimesheet(cmd.Context(), resolvedProjectID)
			}
			if err != nil {
				return err
			}

			// Apply filters
			entries = filterEntries(entries, personStr, sinceDate)

			// Output format
			if formatStr == "json" {
				return outputJSON(entries)
			}

			// Display as table
			return renderEntryTable(entries, false)
		},
	}

	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Project ID")
	cmd.Flags().StringVar(&personStr, "person", "", "Filter by person name (case-insensitive substring match)")
	cmd.Flags().StringVar(&sinceStr, "since", "", "Show entries since (e.g., '7d', '2024-01-01')")
	cmd.Flags().StringVarP(&formatStr, "format", "f", "table", "Output format (table, json)")
	cmd.Flags().Int64Var(&recordingID, "recording", 0, "Filter by recording ID")

	return cmd
}

// filterEntries applies person and date filters
func filterEntries(entries []api.TimesheetEntry, personStr string, sinceDate time.Time) []api.TimesheetEntry {
	var filtered []api.TimesheetEntry

	for _, entry := range entries {
		// Filter by person
		if personStr != "" {
			if !strings.Contains(strings.ToLower(entry.Creator.Name), strings.ToLower(personStr)) {
				continue
			}
		}

		// Filter by date
		if !sinceDate.IsZero() {
			entryDate, err := time.Parse("2006-01-02", entry.Date)
			if err != nil || entryDate.Before(sinceDate) {
				continue
			}
		}

		filtered = append(filtered, entry)
	}

	return filtered
}

// parseSince parses a "since" string (e.g., "7d", "24h", "2024-01-01")
func parseSince(since string) (time.Time, error) {
	// Try parsing as duration (e.g., "7d", "24h")
	if strings.HasSuffix(since, "d") {
		days := since[:len(since)-1]
		d, err := strconv.Atoi(days)
		if err == nil {
			return time.Now().AddDate(0, 0, -d), nil
		}
	}
	if strings.HasSuffix(since, "h") {
		hours := since[:len(since)-1]
		h, err := strconv.Atoi(hours)
		if err == nil {
			return time.Now().Add(-time.Duration(h) * time.Hour), nil
		}
	}

	// Try parsing as ISO date
	t, err := time.Parse("2006-01-02", since)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format: %s (expected Nd, Nh, or YYYY-MM-DD)", since)
	}
	return t, nil
}

// outputJSON outputs entries as JSON
func outputJSON(entries []api.TimesheetEntry) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}

// renderEntryTable displays entries in a table format, optionally with a total summary line
func renderEntryTable(entries []api.TimesheetEntry, showTotal bool) error {
	if len(entries) == 0 {
		fmt.Println("No timesheet entries found")
		return nil
	}

	tp := tableprinter.New(os.Stdout)

	tp.AddField("DATE")
	tp.AddField("HOURS")
	tp.AddField("PERSON")
	tp.AddField("PROJECT")
	tp.AddField("PARENT")
	tp.AddField("DESCRIPTION")
	tp.EndRow()

	var totalHours float64
	for _, entry := range entries {
		tp.AddField(entry.Date)
		tp.AddField(fmt.Sprintf("%.2f", entry.Hours))
		tp.AddField(entry.Creator.Name)
		tp.AddField(entry.Bucket.Name)
		tp.AddField(entry.Parent.Title)

		desc := entry.Description
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}
		tp.AddField(desc)
		tp.EndRow()

		totalHours += entry.Hours
	}

	if err := tp.Render(); err != nil {
		return err
	}

	if showTotal {
		fmt.Printf("\nTotal: %d entries, %.2f hours\n", len(entries), totalHours)
	}

	return nil
}
