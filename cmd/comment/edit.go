package comment

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/markdown"
	"github.com/needmore/bc4/internal/parser"
	"github.com/needmore/bc4/internal/ui"
	"github.com/spf13/cobra"
)

func newEditCmd(f *factory.Factory) *cobra.Command {
	var content string

	cmd := &cobra.Command{
		Use:   "edit <comment-id|url>",
		Short: "Edit an existing comment",
		Long: `Edit an existing comment.

You can provide updated content in several ways:
  - Interactively (default)
  - Via --content flag
  - Via stdin: cat updated.md | bc4 comment edit <comment-id|url>`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get API client from factory
			client, err := f.ApiClient()
			if err != nil {
				return err
			}

			var commentID int64
			var projectID string

			// Parse the argument - could be a URL or ID
			if parser.IsBasecampURL(args[0]) {
				parsed, err := parser.ParseBasecampURL(args[0])
				if err != nil {
					return fmt.Errorf("invalid Basecamp URL: %w", err)
				}
				if parsed.ResourceType != parser.ResourceTypeComment {
					return fmt.Errorf("URL is not a comment URL: %s", args[0])
				}
				commentID = parsed.ResourceID
				projectID = strconv.FormatInt(parsed.ProjectID, 10)
			} else {
				// It's just an ID, we need the project ID from config
				commentID, err = strconv.ParseInt(args[0], 10, 64)
				if err != nil {
					return fmt.Errorf("invalid comment ID: %s", args[0])
				}
				projectID, err = f.ProjectID()
				if err != nil {
					return err
				}
			}

			// Get the existing comment
			comment, err := client.GetComment(f.Context(), projectID, commentID)
			if err != nil {
				return err
			}

			// Check if stdin has data
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				// Data is being piped in
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("failed to read from stdin: %w", err)
				}
				content = strings.TrimSpace(string(data))
			} else if content == "" {
				// No stdin and no flags, use interactive mode
				// Convert existing rich text back to markdown for editing
				converter := markdown.NewConverter()
				existingMarkdown, err := converter.RichTextToMarkdown(comment.Content)
				if err != nil {
					// If conversion fails, use the raw content
					existingMarkdown = comment.Content
				}

				// Edit content
				newContent := existingMarkdown
				if err := huh.NewText().
					Title("Comment content").
					Lines(5).
					Value(&newContent).
					Run(); err != nil {
					return err
				}
				if newContent != existingMarkdown {
					content = newContent
				}
			}

			// Validate that we have changes
			if content == "" {
				fmt.Println("No changes made")
				return nil
			}

			// Convert markdown to rich text
			converter := markdown.NewConverter()
			richContent, err := converter.MarkdownToRichText(content)
			if err != nil {
				return fmt.Errorf("failed to convert markdown: %w", err)
			}

			// Update the comment
			req := api.CommentUpdateRequest{
				Content: richContent,
			}

			updated, err := client.UpdateComment(f.Context(), projectID, commentID, req)
			if err != nil {
				return err
			}

			// Output
			if ui.IsTerminal(os.Stdout) {
				fmt.Printf("✓ Updated comment #%d\n", updated.ID)
			} else {
				fmt.Println(updated.ID)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&content, "content", "c", "", "New comment content (markdown supported)")

	return cmd
}
