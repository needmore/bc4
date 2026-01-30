package checkin

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
	"github.com/needmore/bc4/internal/ui/tableprinter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type answersOptions struct {
	jsonOutput bool
	date       string
	creatorID  int64
}

func newAnswersCmd(f *factory.Factory) *cobra.Command {
	opts := &answersOptions{}

	cmd := &cobra.Command{
		Use:   "answers <question-id|URL>",
		Short: "List answers for a check-in question",
		Long: `List all answers for a specific check-in question.

You can filter answers by date or by person who submitted them.`,
		Example: `  # List all answers for a question
  bc4 checkin answers 12345

  # Filter by date
  bc4 checkin answers 12345 --date 2024-01-15

  # Filter by creator
  bc4 checkin answers 12345 --creator 67890

  # Output as JSON
  bc4 checkin answers 12345 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.jsonOutput = viper.GetBool("json")
			return runAnswers(f, opts, args)
		},
	}

	cmd.Flags().StringVar(&opts.date, "date", "", "Filter by date (YYYY-MM-DD)")
	cmd.Flags().Int64Var(&opts.creatorID, "creator", 0, "Filter by creator person ID")

	return cmd
}

func runAnswers(f *factory.Factory, opts *answersOptions, args []string) error {
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

	// Build list options
	listOpts := &api.AnswerListOptions{
		Date:      opts.date,
		CreatorID: opts.creatorID,
	}

	// Get answers
	answers, err := questionOps.ListAnswers(f.Context(), projectID, parsedID, listOpts)
	if err != nil {
		return fmt.Errorf("failed to list answers: %w", err)
	}

	if opts.jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(answers)
	}

	if len(answers) == 0 {
		fmt.Println("No answers found for this question.")
		return nil
	}

	// Create table
	table := tableprinter.New(os.Stdout)

	// Add headers
	if table.IsTTY() {
		table.AddHeader("ID", "AUTHOR", "CONTENT", "DATE", "CREATED")
	} else {
		table.AddHeader("ID", "AUTHOR", "CONTENT", "GROUP_ON", "CREATED_AT")
	}

	now := time.Now()

	for _, a := range answers {
		// ID
		table.AddIDField(strconv.FormatInt(a.ID, 10), a.Status)

		// Author
		author := ""
		if a.Creator != nil {
			author = a.Creator.Name
		}
		table.AddField(author)

		// Content (truncated)
		content := stripHTML(a.Content)
		if len(content) > 60 {
			content = content[:57] + "..."
		}
		table.AddField(content)

		// Group on date
		table.AddField(a.GroupOn)

		// Created time
		table.AddTimeField(now, a.CreatedAt)

		table.EndRow()
	}

	return table.Render()
}

// stripHTML removes HTML tags from a string (basic implementation)
func stripHTML(s string) string {
	var result []rune
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
		} else if r == '>' {
			inTag = false
		} else if !inTag {
			if r == '\n' || r == '\r' {
				result = append(result, ' ')
			} else {
				result = append(result, r)
			}
		}
	}
	return string(result)
}
