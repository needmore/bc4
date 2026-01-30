package checkin

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type createOptions struct {
	jsonOutput bool
	schedule   string
	days       string
	hour       int
	minute     int
}

func newCreateCmd(f *factory.Factory) *cobra.Command {
	opts := &createOptions{}

	cmd := &cobra.Command{
		Use:   "create <title>",
		Short: "Create a new check-in question",
		Long: `Create a new automated check-in question.

Schedule options:
  every_day          - Ask every day
  every_week         - Ask every week on specified days
  every_other_week   - Ask every other week on specified days
  every_four_weeks   - Ask every four weeks on specified days

Days are specified as comma-separated numbers (0=Sunday through 6=Saturday).
For example: "1,2,3,4,5" for weekdays.`,
		Example: `  # Create a daily check-in at 9am
  bc4 checkin create "What did you accomplish today?" --schedule every_day --hour 9

  # Create a weekly check-in on Mondays at 10am
  bc4 checkin create "What are your goals this week?" --schedule every_week --days 1 --hour 10

  # Create a bi-weekly check-in on Friday afternoons
  bc4 checkin create "Team retrospective" --schedule every_other_week --days 5 --hour 15

  # Output as JSON
  bc4 checkin create "Daily standup" --schedule every_day --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.jsonOutput = viper.GetBool("json")
			return runCreate(f, opts, args)
		},
	}

	cmd.Flags().StringVar(&opts.schedule, "schedule", "every_week", "Schedule frequency: every_day, every_week, every_other_week, every_four_weeks")
	cmd.Flags().StringVar(&opts.days, "days", "1", "Days to ask (comma-separated: 0=Sun, 1=Mon, ..., 6=Sat)")
	cmd.Flags().IntVar(&opts.hour, "hour", 9, "Hour to send reminder (0-23)")
	cmd.Flags().IntVar(&opts.minute, "minute", 0, "Minute to send reminder (0-59)")

	return cmd
}

func runCreate(f *factory.Factory, opts *createOptions, args []string) error {
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

	// Parse days
	days, err := parseDays(opts.days)
	if err != nil {
		return err
	}

	// Validate schedule
	validSchedules := map[string]bool{
		"every_day":        true,
		"every_week":       true,
		"every_other_week": true,
		"every_four_weeks": true,
	}
	if !validSchedules[opts.schedule] {
		return fmt.Errorf("invalid schedule: %s; must be one of: every_day, every_week, every_other_week, every_four_weeks", opts.schedule)
	}

	// Validate hour and minute
	if opts.hour < 0 || opts.hour > 23 {
		return fmt.Errorf("hour must be between 0 and 23")
	}
	if opts.minute < 0 || opts.minute > 59 {
		return fmt.Errorf("minute must be between 0 and 59")
	}

	// Create the question
	req := api.QuestionCreateRequest{
		Title:    args[0],
		Schedule: opts.schedule,
		Days:     days,
		Hour:     opts.hour,
		Minute:   opts.minute,
	}

	question, err := questionOps.CreateQuestion(f.Context(), projectID, questionnaire.ID, req)
	if err != nil {
		return fmt.Errorf("failed to create question: %w", err)
	}

	if opts.jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(question)
	}

	fmt.Printf("Check-in question created successfully (ID: %d)\n", question.ID)
	fmt.Printf("Title: %s\n", question.Title)
	if question.Schedule != nil {
		fmt.Printf("Schedule: %s\n", formatSchedule(question.Schedule))
	}
	if question.AppURL != "" {
		fmt.Printf("URL: %s\n", question.AppURL)
	}

	return nil
}

func parseDays(daysStr string) ([]int, error) {
	if daysStr == "" {
		return nil, nil
	}

	parts := strings.Split(daysStr, ",")
	days := make([]int, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		day, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid day: %s; must be a number 0-6", p)
		}
		if day < 0 || day > 6 {
			return nil, fmt.Errorf("day must be between 0 (Sunday) and 6 (Saturday): got %d", day)
		}
		days = append(days, day)
	}

	return days, nil
}
