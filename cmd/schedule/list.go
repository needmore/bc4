package schedule

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/ui/tableprinter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type listOptions struct {
	jsonOutput bool
}

func newListCmd(f *factory.Factory) *cobra.Command {
	opts := &listOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List schedules in the current project",
		Long: `List all schedules (calendars) available in the current project.

Most projects have a single schedule, but this command shows all available ones.`,
		Example: `  # List all schedules
  bc4 schedule list

  # Output as JSON
  bc4 schedule list --json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.jsonOutput = viper.GetBool("json")
			return runList(f, opts)
		},
	}

	return cmd
}

func runList(f *factory.Factory, opts *listOptions) error {
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

	// Get detailed schedule info if we have an ID
	var scheduleDetail *struct {
		ID           int64
		Title        string
		EntriesCount int
		EntriesURL   string
	}
	if schedule.ID > 0 {
		fullSchedule, err := scheduleOps.GetSchedule(f.Context(), projectID, schedule.ID)
		if err == nil {
			scheduleDetail = &struct {
				ID           int64
				Title        string
				EntriesCount int
				EntriesURL   string
			}{
				ID:           fullSchedule.ID,
				Title:        fullSchedule.Title,
				EntriesCount: fullSchedule.EntriesCount,
				EntriesURL:   fullSchedule.EntriesURL,
			}
		}
	}

	// Output
	if opts.jsonOutput {
		schedules := []interface{}{}
		if scheduleDetail != nil {
			schedules = append(schedules, map[string]interface{}{
				"id":            scheduleDetail.ID,
				"title":         scheduleDetail.Title,
				"entries_count": scheduleDetail.EntriesCount,
			})
		} else {
			schedules = append(schedules, map[string]interface{}{
				"id":    schedule.ID,
				"title": schedule.Title,
			})
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(schedules)
	}

	// Table output
	table := tableprinter.New(os.Stdout)

	table.AddHeader("ID", "TITLE", "ENTRIES")

	if scheduleDetail != nil {
		table.AddField(fmt.Sprintf("%d", scheduleDetail.ID))
		table.AddField(scheduleDetail.Title)
		table.AddField(fmt.Sprintf("%d", scheduleDetail.EntriesCount))
	} else {
		table.AddField(fmt.Sprintf("%d", schedule.ID))
		table.AddField(schedule.Title)
		table.AddField("-")
	}
	table.EndRow()

	return table.Render()
}
