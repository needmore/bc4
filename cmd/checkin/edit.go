package checkin

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type editOptions struct {
	jsonOutput bool
	title      string
	schedule   string
	days       string
	hour       int
	minute     int
	hourSet    bool
	minuteSet  bool
}

func newEditCmd(f *factory.Factory) *cobra.Command {
	opts := &editOptions{}

	cmd := &cobra.Command{
		Use:   "edit <question-id|URL>",
		Short: "Edit a check-in question",
		Long: `Edit an existing check-in question's title or schedule.

Only the fields you specify will be updated.`,
		Example: `  # Update the title
  bc4 checkin edit 12345 --title "New question text"

  # Update the schedule
  bc4 checkin edit 12345 --schedule every_week --days 1,3,5

  # Update the time
  bc4 checkin edit 12345 --hour 14 --minute 30

  # Update multiple fields
  bc4 checkin edit 12345 --title "Weekly update" --schedule every_week --days 1 --hour 9`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.jsonOutput = viper.GetBool("json")
			opts.hourSet = cmd.Flags().Changed("hour")
			opts.minuteSet = cmd.Flags().Changed("minute")
			return runEdit(f, opts, args)
		},
	}

	cmd.Flags().StringVar(&opts.title, "title", "", "New question title")
	cmd.Flags().StringVar(&opts.schedule, "schedule", "", "Schedule frequency: every_day, every_week, every_other_week, every_four_weeks")
	cmd.Flags().StringVar(&opts.days, "days", "", "Days to ask (comma-separated: 0=Sun, 1=Mon, ..., 6=Sat)")
	cmd.Flags().IntVar(&opts.hour, "hour", 0, "Hour to send reminder (0-23)")
	cmd.Flags().IntVar(&opts.minute, "minute", 0, "Minute to send reminder (0-59)")

	return cmd
}

func runEdit(f *factory.Factory, opts *editOptions, args []string) error {
	client, err := f.ApiClient()
	if err != nil {
		return err
	}
	questionOps := client.Questions()

	projectID, err := f.ProjectID()
	if err != nil {
		return err
	}

	// Parse question ID (could be numeric ID or URL)
	parsedID, parsedURL, err := parser.ParseArgument(args[0])
	if err != nil {
		return fmt.Errorf("invalid question ID or URL: %s", args[0])
	}

	if parsedURL != nil {
		if parsedURL.AccountID > 0 {
			f = f.WithAccount(strconv.FormatInt(parsedURL.AccountID, 10))
		}
		if parsedURL.ProjectID > 0 {
			projectID = strconv.FormatInt(parsedURL.ProjectID, 10)
		}
	}

	// Build update request
	req := api.QuestionUpdateRequest{}
	hasUpdates := false

	if opts.title != "" {
		req.Title = &opts.title
		hasUpdates = true
	}

	if opts.schedule != "" {
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
		req.Schedule = &opts.schedule
		hasUpdates = true
	}

	if opts.days != "" {
		days, err := parseDays(opts.days)
		if err != nil {
			return err
		}
		req.Days = days
		hasUpdates = true
	}

	if opts.hourSet {
		if opts.hour < 0 || opts.hour > 23 {
			return fmt.Errorf("hour must be between 0 and 23")
		}
		req.Hour = &opts.hour
		hasUpdates = true
	}

	if opts.minuteSet {
		if opts.minute < 0 || opts.minute > 59 {
			return fmt.Errorf("minute must be between 0 and 59")
		}
		req.Minute = &opts.minute
		hasUpdates = true
	}

	if !hasUpdates {
		return fmt.Errorf("no changes specified; use --title, --schedule, --days, --hour, or --minute")
	}

	// Update the question
	question, err := questionOps.UpdateQuestion(f.Context(), projectID, parsedID, req)
	if err != nil {
		return fmt.Errorf("failed to update question: %w", err)
	}

	if opts.jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(question)
	}

	fmt.Printf("Check-in question %d updated successfully.\n", question.ID)
	fmt.Printf("Title: %s\n", question.Title)
	if question.Schedule != nil {
		fmt.Printf("Schedule: %s\n", formatSchedule(question.Schedule))
	}

	return nil
}
