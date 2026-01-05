package comment

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/attachments"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/markdown"
	"github.com/needmore/bc4/internal/parser"
	"github.com/needmore/bc4/internal/ui"
	"github.com/spf13/cobra"
)

func newAttachCmd(f *factory.Factory) *cobra.Command {
	var attachmentPath string
	var commentIDFlag int64

	cmd := &cobra.Command{
		Use:   "attach <recording-id|url>",
		Short: "Append an attachment to an existing comment",
		Long: `Upload a file and append it to an existing comment's body.

If no comment ID is provided, the latest comment on the recording is used.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if attachmentPath == "" {
				return fmt.Errorf("--attach is required")
			}

			client, err := f.ApiClient()
			if err != nil {
				return err
			}

			var projectID string
			var recordingID int64
			var targetCommentID int64

			if parser.IsBasecampURL(args[0]) {
				parsed, err := parser.ParseBasecampURL(args[0])
				if err != nil {
					return fmt.Errorf("invalid Basecamp URL: %w", err)
				}
				projectID = strconv.FormatInt(parsed.ProjectID, 10)
				if parsed.ResourceType == parser.ResourceTypeComment {
					targetCommentID = parsed.ResourceID
				} else {
					recordingID = parsed.ResourceID
				}
			} else {
				// Treat argument as recording ID
				recordingID, err = strconv.ParseInt(args[0], 10, 64)
				if err != nil {
					return fmt.Errorf("invalid recording ID: %s", args[0])
				}
				projectID, err = f.ProjectID()
				if err != nil {
					return err
				}
			}

			// Allow override via flag
			if commentIDFlag != 0 {
				targetCommentID = commentIDFlag
			}

			// If no target comment, pick latest from recording
			if targetCommentID == 0 {
				if recordingID == 0 {
					return fmt.Errorf("comment ID or recording ID is required")
				}
				comments, err := client.ListComments(f.Context(), projectID, recordingID)
				if err != nil {
					return err
				}
				if len(comments) == 0 {
					return fmt.Errorf("no comments found for recording %d", recordingID)
				}
				targetCommentID = newestCommentID(comments)
			}

			existing, err := client.GetComment(f.Context(), projectID, targetCommentID)
			if err != nil {
				return err
			}

			fileData, err := os.ReadFile(attachmentPath)
			if err != nil {
				return fmt.Errorf("failed to read attachment: %w", err)
			}

			filename := filepath.Base(attachmentPath)
			upload, err := client.UploadAttachment(filename, fileData, "")
			if err != nil {
				return fmt.Errorf("failed to upload attachment: %w", err)
			}

			newContent := existing.Content + attachments.BuildTag(upload.AttachableSGID)

			converter := markdown.NewConverter()
			if err := converter.ValidateBasecampHTML(newContent); err != nil {
				return fmt.Errorf("resulting content is invalid: %w", err)
			}

			updated, err := client.UpdateComment(f.Context(), projectID, targetCommentID, api.CommentUpdateRequest{
				Content: newContent,
			})
			if err != nil {
				return err
			}

			if ui.IsTerminal(os.Stdout) {
				fmt.Printf("✓ Attached file to comment #%d\n", updated.ID)
			} else {
				fmt.Println(updated.ID)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&attachmentPath, "attach", "", "Path to file to attach")
	cmd.Flags().Int64Var(&commentIDFlag, "comment-id", 0, "Append to a specific comment ID instead of the latest")

	return cmd
}

func newestCommentID(comments []api.Comment) int64 {
	var latest api.Comment
	var hasLatest bool

	for _, c := range comments {
		if !hasLatest || c.CreatedAt.After(latest.CreatedAt) || (c.CreatedAt.Equal(latest.CreatedAt) && c.ID > latest.ID) {
			latest = c
			hasLatest = true
		}
	}

	return latest.ID
}
