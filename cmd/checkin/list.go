package checkin

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/needmore/bc4/internal/api"
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
		Short: "List check-in questions in a project",
		Long: `List all automated check-in questions in the current project.

Shows the question title, schedule, paused status, and answer count.`,
		Example: `  # List check-ins in the current project
  bc4 checkin list

  # Output as JSON
  bc4 checkin list --json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Apply overrides from persistent flags
			if accountID := viper.GetString("account"); accountID != "" {
				f = f.WithAccount(accountID)
			}
			if projectID := viper.GetString("project"); projectID != "" {
				f = f.WithProject(projectID)
			}

			opts.jsonOutput = viper.GetBool("json")
			return runList(f, opts)
		},
	}

	return cmd
}

func runList(f *factory.Factory, opts *listOptions) error {
	client, err := f.ApiClient()
	if err != nil {
		return err
	}
	questionOps := client.Questions()

	projectID, err := f.ProjectID()
	if err != nil {
		return err
	}

	// Get the questionnaire for this project
	questionnaire, err := questionOps.GetProjectQuestionnaire(f.Context(), projectID)
	if err != nil {
		return fmt.Errorf("failed to get questionnaire: %w", err)
	}

	// List all questions
	questions, err := questionOps.ListQuestions(f.Context(), projectID, questionnaire.ID)
	if err != nil {
		return fmt.Errorf("failed to list questions: %w", err)
	}

	if opts.jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(questions)
	}

	if len(questions) == 0 {
		fmt.Println("No check-in questions found in this project.")
		return nil
	}

	// Create table
	table := tableprinter.New(os.Stdout)

	// Add headers
	if table.IsTTY() {
		table.AddHeader("ID", "QUESTION", "SCHEDULE", "STATUS", "ANSWERS", "UPDATED")
	} else {
		table.AddHeader("ID", "QUESTION", "SCHEDULE", "PAUSED", "ANSWERS", "UPDATED")
	}

	now := time.Now()

	for _, q := range questions {
		// ID
		table.AddIDField(strconv.FormatInt(q.ID, 10), q.Status)

		// Question title
		title := q.Title
		if len(title) > 50 {
			title = title[:47] + "..."
		}
		table.AddField(title)

		// Schedule
		schedule := formatSchedule(q.Schedule)
		table.AddField(schedule)

		// Paused status
		if table.IsTTY() {
			if q.Paused {
				table.AddColorField("paused", "trashed")
			} else {
				table.AddColorField("active", "active")
			}
		} else {
			table.AddField(strconv.FormatBool(q.Paused))
		}

		// Answer count
		table.AddField(strconv.Itoa(q.AnswersCount))

		// Updated time
		table.AddTimeField(now, q.UpdatedAt)

		table.EndRow()
	}

	return table.Render()
}

func formatSchedule(s *api.QuestionSchedule) string {
	if s == nil {
		return "unknown"
	}

	freq := s.Frequency
	switch freq {
	case "every_day":
		return fmt.Sprintf("daily @ %02d:%02d", s.Hour, s.Minute)
	case "every_week":
		days := formatDays(s.Days)
		return fmt.Sprintf("weekly %s @ %02d:%02d", days, s.Hour, s.Minute)
	case "every_other_week":
		days := formatDays(s.Days)
		return fmt.Sprintf("bi-weekly %s @ %02d:%02d", days, s.Hour, s.Minute)
	case "every_four_weeks":
		days := formatDays(s.Days)
		return fmt.Sprintf("monthly %s @ %02d:%02d", days, s.Hour, s.Minute)
	default:
		return freq
	}
}

func formatDays(days []int) string {
	if len(days) == 0 {
		return ""
	}

	dayNames := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	var result []string
	for _, d := range days {
		if d >= 0 && d <= 6 {
			result = append(result, dayNames[d])
		}
	}

	if len(result) == 5 && !containsDay(days, 0) && !containsDay(days, 6) {
		return "weekdays"
	}

	if len(result) == 1 {
		return result[0]
	}

	if len(result) > 3 {
		return fmt.Sprintf("%d days", len(result))
	}

	// Join with commas
	joined := ""
	for i, r := range result {
		if i > 0 {
			joined += ","
		}
		joined += r
	}
	return joined
}

func containsDay(days []int, day int) bool {
	for _, d := range days {
		if d == day {
			return true
		}
	}
	return false
}
