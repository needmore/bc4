package message

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/cmdutil"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/markdown"
	"github.com/needmore/bc4/internal/parser"
	"github.com/needmore/bc4/internal/ui"
	"github.com/spf13/cobra"
)

func newEditCmd(f *factory.Factory) *cobra.Command {
	var (
		title      string
		content    string
		categoryID int64
	)

	cmd := &cobra.Command{
		Use:   "edit <message-id|url>",
		Short: "Edit an existing message",
		Long: `Edit an existing message.

You can provide updated content in several ways:
  - Interactively (default)
  - Via --content flag
  - Via stdin: cat updated.md | bc4 message edit <message-id>`,
		Example: `bc4 message edit 12345
bc4 message edit 12345 --title "New Title"
cat updated.md | bc4 message edit 12345`,
		Args: cmdutil.ExactArgs(1, "message-id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get API client from factory
			client, err := f.ApiClient()
			if err != nil {
				return err
			}

			var messageID int64
			var projectID string

			// Parse the argument - could be a URL or ID
			if parser.IsBasecampURL(args[0]) {
				parsed, err := parser.ParseBasecampURL(args[0])
				if err != nil {
					return fmt.Errorf("invalid Basecamp URL: %w", err)
				}
				if parsed.ResourceType != parser.ResourceTypeMessage {
					return fmt.Errorf("URL is not a message URL: %s", args[0])
				}
				messageID = parsed.ResourceID
				projectID = strconv.FormatInt(parsed.ProjectID, 10)
			} else {
				// It's just an ID, we need the project ID from config
				messageID, err = strconv.ParseInt(args[0], 10, 64)
				if err != nil {
					return fmt.Errorf("invalid message ID: %s", args[0])
				}
				projectID, err = f.ProjectID()
				if err != nil {
					return err
				}
			}

			// Get the existing message
			message, err := client.GetMessage(f.Context(), projectID, messageID)
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
			} else if content == "" && title == "" {
				// No stdin and no flags, use interactive mode
				// Convert existing rich text back to markdown for editing
				converter := markdown.NewConverter()
				existingMarkdown, err := converter.RichTextToMarkdown(message.Content)
				if err != nil {
					// If conversion fails, use the raw content
					existingMarkdown = message.Content
				}

				// Edit title
				newTitle := message.Subject
				if err := huh.NewInput().
					Title("Message subject").
					Value(&newTitle).
					Run(); err != nil {
					return err
				}
				if newTitle != message.Subject {
					title = newTitle
				}

				// Edit content
				newContent := existingMarkdown
				if err := huh.NewText().
					Title("Message content").
					Lines(10).
					Value(&newContent).
					Run(); err != nil {
					return err
				}
				if newContent != existingMarkdown {
					content = newContent
				}
			}

			// Build update request
			req := api.MessageUpdateRequest{}
			hasUpdate := false

			if title != "" {
				req.Subject = title
				hasUpdate = true
			}

			if content != "" {
				// Convert markdown to rich text
				converter := markdown.NewConverter()
				richContent, err := converter.MarkdownToRichText(content)
				if err != nil {
					return fmt.Errorf("failed to convert markdown: %w", err)
				}
				req.Content = richContent
				hasUpdate = true
			}

			if categoryID > 0 {
				req.CategoryID = &categoryID
				hasUpdate = true
			}

			if !hasUpdate {
				fmt.Println("No changes made")
				return nil
			}

			// Update the message
			updated, err := client.UpdateMessage(f.Context(), projectID, messageID, req)
			if err != nil {
				return err
			}

			// Output
			if ui.IsTerminal(os.Stdout) {
				fmt.Printf("✓ Updated message #%d: %s\n", updated.ID, updated.Subject)
			} else {
				fmt.Println(updated.ID)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&title, "title", "t", "", "New message subject")
	cmd.Flags().StringVarP(&content, "content", "c", "", "New message content (markdown supported)")
	cmd.Flags().Int64Var(&categoryID, "category-id", 0, "Category ID")

	return cmd
}
