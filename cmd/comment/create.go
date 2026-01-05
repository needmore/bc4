package comment

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/attachments"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/markdown"
	"github.com/needmore/bc4/internal/parser"
	"github.com/needmore/bc4/internal/ui"
	"github.com/spf13/cobra"
)

func newCreateCmd(f *factory.Factory) *cobra.Command {
	var content string
	var attachmentPath string

	cmd := &cobra.Command{
		Use:   "create <recording-id|url>",
		Short: "Create a comment",
		Long: `Create a new comment on a Basecamp recording (todo, message, document, or card).

You can provide comment content in several ways:
  - Interactively (default)
  - Via --content flag
  - Via stdin: echo "content" | bc4 comment create <recording-id|url>
  - From file: cat comment.md | bc4 comment create <recording-id|url>`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get API client from factory
			client, err := f.ApiClient()
			if err != nil {
				return err
			}

			var recordingID int64
			var projectID string

			// Parse the argument - could be a URL or ID
			if parser.IsBasecampURL(args[0]) {
				parsed, err := parser.ParseBasecampURL(args[0])
				if err != nil {
					return fmt.Errorf("invalid Basecamp URL: %w", err)
				}
				recordingID = parsed.ResourceID
				projectID = strconv.FormatInt(parsed.ProjectID, 10)
			} else {
				// It's just an ID, we need the project ID from config
				recordingID, err = strconv.ParseInt(args[0], 10, 64)
				if err != nil {
					return fmt.Errorf("invalid recording ID: %s", args[0])
				}
				projectID, err = f.ProjectID()
				if err != nil {
					return err
				}
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
				// No stdin and no content flag, use interactive mode
				if err := huh.NewText().
					Title("Comment content").
					Placeholder("Write your comment in Markdown...").
					Lines(5).
					Value(&content).
					Run(); err != nil {
					return err
				}
			}

			// Validate content (allow empty if attachment is provided)
			if content == "" && attachmentPath == "" {
				return fmt.Errorf("comment content or --attach is required")
			}

			// Convert markdown to rich text when present
			var richContent string
			if strings.TrimSpace(content) != "" {
				converter := markdown.NewConverter()
				rc, err := converter.MarkdownToRichText(content)
				if err != nil {
					return fmt.Errorf("failed to convert markdown: %w", err)
				}
				richContent = rc
			}

			// Attach file if provided
			if attachmentPath != "" {
				fileData, err := os.ReadFile(attachmentPath)
				if err != nil {
					return fmt.Errorf("failed to read attachment: %w", err)
				}
				filename := filepath.Base(attachmentPath)
				upload, err := client.UploadAttachment(filename, fileData, "")
				if err != nil {
					return fmt.Errorf("failed to upload attachment: %w", err)
				}

				tag := attachments.BuildTag(upload.AttachableSGID)
				richContent += tag
			}

			// Create the comment
			req := api.CommentCreateRequest{
				Content: richContent,
			}

			comment, err := client.CreateComment(f.Context(), projectID, recordingID, req)
			if err != nil {
				return err
			}

			// Output
			if ui.IsTerminal(os.Stdout) {
				fmt.Printf("✓ Created comment #%d\n", comment.ID)
			} else {
				fmt.Println(comment.ID)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&content, "content", "", "Comment content (Markdown)")
	cmd.Flags().StringVar(&attachmentPath, "attach", "", "Path to file to attach to the comment")

	return cmd
}
