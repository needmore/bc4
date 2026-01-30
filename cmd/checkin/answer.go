package checkin

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/markdown"
	"github.com/needmore/bc4/internal/parser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type answerOptions struct {
	jsonOutput bool
	file       string
	markdown   bool
}

func newAnswerCmd(f *factory.Factory) *cobra.Command {
	opts := &answerOptions{}

	cmd := &cobra.Command{
		Use:   "answer <question-id|URL> [content]",
		Short: "Post an answer to a check-in question",
		Long: `Post an answer to a check-in question.

Content can be provided as:
- A command-line argument
- From a file using --file
- From stdin (pipe content to bc4)

If --markdown is specified (or --md), content will be converted from Markdown to HTML.`,
		Example: `  # Post a simple answer
  bc4 checkin answer 12345 "Today I worked on the new feature"

  # Post from a file
  bc4 checkin answer 12345 --file update.txt

  # Post from stdin
  echo "My update" | bc4 checkin answer 12345

  # Post markdown content
  bc4 checkin answer 12345 "## Summary\n- Fixed bugs\n- Added tests" --markdown`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.jsonOutput = viper.GetBool("json")
			return runAnswer(f, opts, args)
		},
	}

	cmd.Flags().StringVarP(&opts.file, "file", "f", "", "Read content from file")
	cmd.Flags().BoolVar(&opts.markdown, "markdown", false, "Convert markdown to HTML")
	cmd.Flags().BoolVar(&opts.markdown, "md", false, "Convert markdown to HTML (alias)")

	return cmd
}

func runAnswer(f *factory.Factory, opts *answerOptions, args []string) error {
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

	// Get the content
	content, err := getAnswerContent(opts, args)
	if err != nil {
		return err
	}

	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("answer content cannot be empty")
	}

	// Convert markdown if requested
	if opts.markdown {
		converter := markdown.NewConverter()
		htmlContent, err := converter.MarkdownToRichText(content)
		if err != nil {
			return fmt.Errorf("failed to convert markdown: %w", err)
		}
		content = htmlContent
	}

	// Create the answer
	req := api.AnswerCreateRequest{
		Content: content,
	}

	answer, err := questionOps.CreateAnswer(f.Context(), projectID, parsedID, req)
	if err != nil {
		return fmt.Errorf("failed to create answer: %w", err)
	}

	if opts.jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(answer)
	}

	fmt.Printf("Answer posted successfully (ID: %d)\n", answer.ID)
	if answer.AppURL != "" {
		fmt.Printf("URL: %s\n", answer.AppURL)
	}

	return nil
}

func getAnswerContent(opts *answerOptions, args []string) (string, error) {
	// Check for file input
	if opts.file != "" {
		data, err := os.ReadFile(opts.file)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}
		return string(data), nil
	}

	// Check for command-line argument
	if len(args) >= 2 {
		return args[1], nil
	}

	// Check for stdin
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Data is being piped to stdin
		reader := bufio.NewReader(os.Stdin)
		data, err := io.ReadAll(reader)
		if err != nil {
			return "", fmt.Errorf("failed to read stdin: %w", err)
		}
		return string(data), nil
	}

	return "", fmt.Errorf("no content provided; use argument, --file, or pipe content to stdin")
}
