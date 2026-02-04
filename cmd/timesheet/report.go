package timesheet

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/ui/tableprinter"
	"github.com/spf13/cobra"
)

func newReportCmd(f *factory.Factory) *cobra.Command {
	var (
		accountID string
		projectID string
		personStr string
		startDate string
		endDate   string
		formatStr string
		groupBy   string
	)

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate account-wide timesheet report",
		Long: `Generate a timesheet report across all projects in the account.

This command fetches timesheet entries from all accessible projects and provides
options to filter by date range, person, or project. The report can be grouped
by person or project for easier analysis.

Note: Without date filters, the API returns entries from the last month by default.`,
		Example: `  bc4 timesheet report                         # Last month's entries
  bc4 timesheet report --start 2024-01-01      # From start of year
  bc4 timesheet report --start 2024-01-01 --end 2024-01-31
  bc4 timesheet report --person "john"         # Filter by person
  bc4 timesheet report --project 123456        # Filter by project
  bc4 timesheet report --group-by person       # Group by person
  bc4 timesheet report --group-by project      # Group by project`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Apply flag overrides
			if accountID != "" {
				f = f.WithAccount(accountID)
			}

			// Get API client from factory
			client, err := f.ApiClient()
			if err != nil {
				return err
			}

			// Build options
			opts := &api.TimesheetReportOptions{}

			// Parse date filters
			if startDate != "" {
				start, err := time.Parse("2006-01-02", startDate)
				if err != nil {
					return fmt.Errorf("invalid start date format (expected YYYY-MM-DD): %w", err)
				}
				opts.StartDate = &start
			}
			if endDate != "" {
				end, err := time.Parse("2006-01-02", endDate)
				if err != nil {
					return fmt.Errorf("invalid end date format (expected YYYY-MM-DD): %w", err)
				}
				opts.EndDate = &end
			}

			// Parse person filter
			if personStr != "" {
				// We need to convert name to ID - fetch all people first
				people, err := client.GetAllPeople(cmd.Context())
				if err != nil {
					return fmt.Errorf("failed to fetch people: %w", err)
				}
				// Find matching person
				var personID int64
				for _, p := range people {
					if strings.Contains(strings.ToLower(p.Name), strings.ToLower(personStr)) {
						personID = p.ID
						break
					}
				}
				if personID == 0 {
					return fmt.Errorf("no person found matching: %s", personStr)
				}
				opts.PersonID = &personID
			}

			// Parse project filter
			if projectID != "" {
				opts.BucketID = &projectID
			}

			// Fetch report
			entries, err := client.GetTimesheetReport(cmd.Context(), opts)
			if err != nil {
				return err
			}

			// Output format
			if formatStr == "json" {
				return outputJSON(entries)
			}

			// Display based on grouping
			switch groupBy {
			case "person":
				return displayGroupedByPerson(entries)
			case "project":
				return displayGroupedByProject(entries)
			default:
				return displayReportTable(entries)
			}
		},
	}

	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Filter by project ID")
	cmd.Flags().StringVar(&personStr, "person", "", "Filter by person name (case-insensitive substring match)")
	cmd.Flags().StringVar(&startDate, "start", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&endDate, "end", "", "End date (YYYY-MM-DD)")
	cmd.Flags().StringVarP(&formatStr, "format", "f", "table", "Output format (table, json)")
	cmd.Flags().StringVar(&groupBy, "group-by", "", "Group results by (person, project)")

	return cmd
}

// displayReportTable displays the report as a flat table
func displayReportTable(entries []api.TimesheetEntry) error {
	if len(entries) == 0 {
		fmt.Println("No timesheet entries found")
		return nil
	}

	// Create table
	tp := tableprinter.New(os.Stdout)

	// Add headers
	tp.AddField("DATE")
	tp.AddField("HOURS")
	tp.AddField("PERSON")
	tp.AddField("PROJECT")
	tp.AddField("PARENT")
	tp.AddField("DESCRIPTION")
	tp.EndRow()

	// Calculate total hours
	var totalHours float64

	// Add rows
	for _, entry := range entries {
		tp.AddField(entry.Date)
		tp.AddField(fmt.Sprintf("%.2f", entry.Hours))
		tp.AddField(entry.Creator.Name)
		tp.AddField(entry.Bucket.Name)
		tp.AddField(entry.Parent.Title)

		// Truncate description if too long
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

	// Print summary
	fmt.Printf("\nTotal: %d entries, %.2f hours\n", len(entries), totalHours)

	return nil
}

// displayGroupedByPerson groups entries by person and shows totals
func displayGroupedByPerson(entries []api.TimesheetEntry) error {
	if len(entries) == 0 {
		fmt.Println("No timesheet entries found")
		return nil
	}

	// Group by person
	grouped := make(map[string]struct {
		name  string
		hours float64
		count int
	})

	for _, entry := range entries {
		key := strconv.FormatInt(entry.Creator.ID, 10)
		g := grouped[key]
		g.name = entry.Creator.Name
		g.hours += entry.Hours
		g.count++
		grouped[key] = g
	}

	// Create table
	tp := tableprinter.New(os.Stdout)
	tp.AddField("PERSON")
	tp.AddField("ENTRIES")
	tp.AddField("TOTAL HOURS")
	tp.AddField("AVG HOURS")
	tp.EndRow()

	var totalHours float64
	var totalCount int

	for _, g := range grouped {
		tp.AddField(g.name)
		tp.AddField(fmt.Sprintf("%d", g.count))
		tp.AddField(fmt.Sprintf("%.2f", g.hours))
		tp.AddField(fmt.Sprintf("%.2f", g.hours/float64(g.count)))
		tp.EndRow()

		totalHours += g.hours
		totalCount += g.count
	}

	if err := tp.Render(); err != nil {
		return err
	}

	fmt.Printf("\nTotal: %d entries, %.2f hours across %d people\n", totalCount, totalHours, len(grouped))

	return nil
}

// displayGroupedByProject groups entries by project and shows totals
func displayGroupedByProject(entries []api.TimesheetEntry) error {
	if len(entries) == 0 {
		fmt.Println("No timesheet entries found")
		return nil
	}

	// Group by project
	grouped := make(map[string]struct {
		name  string
		hours float64
		count int
	})

	for _, entry := range entries {
		key := strconv.FormatInt(entry.Bucket.ID, 10)
		g := grouped[key]
		g.name = entry.Bucket.Name
		g.hours += entry.Hours
		g.count++
		grouped[key] = g
	}

	// Create table
	tp := tableprinter.New(os.Stdout)
	tp.AddField("PROJECT")
	tp.AddField("ENTRIES")
	tp.AddField("TOTAL HOURS")
	tp.AddField("AVG HOURS")
	tp.EndRow()

	var totalHours float64
	var totalCount int

	for _, g := range grouped {
		tp.AddField(g.name)
		tp.AddField(fmt.Sprintf("%d", g.count))
		tp.AddField(fmt.Sprintf("%.2f", g.hours))
		tp.AddField(fmt.Sprintf("%.2f", g.hours/float64(g.count)))
		tp.EndRow()

		totalHours += g.hours
		totalCount += g.count
	}

	if err := tp.Render(); err != nil {
		return err
	}

	fmt.Printf("\nTotal: %d entries, %.2f hours across %d projects\n", totalCount, totalHours, len(grouped))

	return nil
}
