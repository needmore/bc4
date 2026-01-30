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

type viewOptions struct {
	jsonOutput bool
}

func newViewCmd(f *factory.Factory) *cobra.Command {
	opts := &viewOptions{}

	cmd := &cobra.Command{
		Use:   "view <question-id|URL>",
		Short: "View a specific check-in question",
		Long: `View details of a specific check-in question.

You can specify the question using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL`,
		Example: `  # View a check-in question by ID
  bc4 checkin view 12345

  # View from URL
  bc4 checkin view "https://3.basecamp.com/123/buckets/456/questions/789"

  # Output as JSON
  bc4 checkin view 12345 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.jsonOutput = viper.GetBool("json")
			return runView(f, opts, args)
		},
	}

	return cmd
}

func runView(f *factory.Factory, opts *viewOptions, args []string) error {
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

	// Get the question
	question, err := questionOps.GetQuestion(f.Context(), projectID, parsedID)
	if err != nil {
		return fmt.Errorf("failed to get question: %w", err)
	}

	if opts.jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(question)
	}

	// Human-readable output
	printQuestion(question)

	return nil
}

func printQuestion(q *api.Question) {
	fmt.Printf("Check-in: %s\n", q.Title)
	fmt.Printf("ID: %d\n", q.ID)

	if q.Schedule != nil {
		fmt.Printf("Schedule: %s\n", formatSchedule(q.Schedule))
	}

	if q.Paused {
		fmt.Println("Status: PAUSED")
	} else {
		fmt.Println("Status: Active")
	}

	fmt.Printf("Answers: %d\n", q.AnswersCount)

	if q.Creator != nil {
		fmt.Printf("Created by: %s\n", q.Creator.Name)
	}

	if !q.CreatedAt.IsZero() {
		fmt.Printf("Created: %s\n", q.CreatedAt.Format("2006-01-02 15:04"))
	}

	if q.AppURL != "" {
		fmt.Printf("URL: %s\n", q.AppURL)
	}

	if q.NotificationSettings != nil {
		fmt.Println()
		fmt.Println("Notification Settings:")
		fmt.Printf("  Responding: %v\n", q.NotificationSettings.Responding)
		fmt.Printf("  Subscribed: %v\n", q.NotificationSettings.Subscribed)
	}
}
