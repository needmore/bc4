package document

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
	"github.com/needmore/bc4/internal/mentions"
	"github.com/needmore/bc4/internal/parser"
	"github.com/needmore/bc4/internal/ui"
	"github.com/spf13/cobra"
)

func newEditCmd(f *factory.Factory) *cobra.Command {
	var (
		title   string
		content string
	)

	cmd := &cobra.Command{
		Use:   "edit [document-id|url]",
		Short: "Edit an existing document",
		Long: `Edit an existing document.

You can provide updated content in several ways:
  - Interactively (default)
  - Via --content flag
  - Via stdin: cat updated.md | bc4 document edit [document-id]`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get API client from factory
			client, err := f.ApiClient()
			if err != nil {
				return err
			}

			var documentID int64
			var projectID string

			// Parse the argument - could be a URL or ID
			if parser.IsBasecampURL(args[0]) {
				parsed, err := parser.ParseBasecampURL(args[0])
				if err != nil {
					return fmt.Errorf("invalid Basecamp URL: %w", err)
				}
				if parsed.ResourceType != parser.ResourceTypeDocument {
					return fmt.Errorf("URL is not a document URL: %s", args[0])
				}
				documentID = parsed.ResourceID
				projectID = strconv.FormatInt(parsed.ProjectID, 10)
			} else {
				// It's just an ID, we need the project ID from config
				documentID, err = strconv.ParseInt(args[0], 10, 64)
				if err != nil {
					return fmt.Errorf("invalid document ID: %s", args[0])
				}
				projectID, err = f.ProjectID()
				if err != nil {
					return err
				}
			}

			// Get the existing document
			document, err := client.GetDocument(f.Context(), projectID, documentID)
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
				existingMarkdown, err := converter.RichTextToMarkdown(document.Content)
				if err != nil {
					// If conversion fails, use the raw content
					existingMarkdown = document.Content
				}

				// Edit title
				newTitle := document.Title
				if err := huh.NewInput().
					Title("Document title").
					Value(&newTitle).
					Run(); err != nil {
					return err
				}
				if newTitle != document.Title {
					title = newTitle
				}

				// Edit content
				newContent := existingMarkdown
				if err := huh.NewText().
					Title("Document content").
					Lines(15).
					Value(&newContent).
					Run(); err != nil {
					return err
				}
				if newContent != existingMarkdown {
					content = newContent
				}
			}

			// Build update request
			req := api.DocumentUpdateRequest{}
			hasUpdate := false

			if title != "" {
				req.Title = title
				hasUpdate = true
			}

			if content != "" {
				// Convert markdown to rich text
				converter := markdown.NewConverter()
				richContent, err := converter.MarkdownToRichText(content)
				if err != nil {
					return fmt.Errorf("failed to convert markdown: %w", err)
				}

				// Replace inline @Name mentions with bc-attachment tags
				richContent, err = mentions.Resolve(f.Context(), richContent, client.Client, projectID)
				if err != nil {
					return fmt.Errorf("failed to resolve mentions: %w", err)
				}

				req.Content = richContent
				hasUpdate = true
			}

			if !hasUpdate {
				fmt.Println("No changes made")
				return nil
			}

			// Update the document
			updated, err := client.UpdateDocument(f.Context(), projectID, documentID, req)
			if err != nil {
				return err
			}

			// Output
			if ui.IsTerminal(os.Stdout) {
				fmt.Printf("✓ Updated document #%d: %s\n", updated.ID, updated.Title)
			} else {
				fmt.Println(updated.ID)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&title, "title", "t", "", "New document title")
	cmd.Flags().StringVarP(&content, "content", "c", "", "New document content (markdown supported)")

	return cmd
}
