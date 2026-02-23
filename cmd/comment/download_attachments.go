package comment

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/needmore/bc4/internal/download"
	"github.com/needmore/bc4/internal/factory"
	"github.com/needmore/bc4/internal/parser"
)

func newDownloadAttachmentsCmd(f *factory.Factory) *cobra.Command {
	var accountID string
	var projectID string
	var outputDir string
	var overwrite bool
	var attachmentIndex int

	cmd := &cobra.Command{
		Use:   "download-attachments [comment-id or URL]",
		Short: "Download attachments from a comment",
		Long: `Download all attachments (images and files) from a comment to local files.

This command fetches attachment metadata from the Basecamp API and downloads
the actual files using OAuth authentication. You can download all attachments
or select specific ones.

You can specify the comment using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/comments/12345")`,
		Example: `  # Download all attachments from a comment
  bc4 comment download-attachments 123456

  # Download to specific directory
  bc4 comment download-attachments 123456 --output-dir ~/Downloads

  # Download only the first attachment
  bc4 comment download-attachments 123456 --attachment 1

  # Overwrite existing files
  bc4 comment download-attachments 123456 --overwrite

  # Using Basecamp URL
  bc4 comment download-attachments https://3.basecamp.com/123/buckets/456/comments/789`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if accountID != "" {
				f = f.WithAccount(accountID)
			}
			if projectID != "" {
				f = f.WithProject(projectID)
			}

			commentID, parsedURL, err := parser.ParseArgument(args[0])
			if err != nil {
				return fmt.Errorf("invalid comment ID or URL: %s", args[0])
			}

			var bucketID string
			if parsedURL != nil {
				if parsedURL.ResourceType != parser.ResourceTypeComment {
					return fmt.Errorf("URL is not for a comment: %s", args[0])
				}
				if parsedURL.AccountID > 0 {
					f = f.WithAccount(strconv.FormatInt(parsedURL.AccountID, 10))
				}
				if parsedURL.ProjectID > 0 {
					f = f.WithProject(strconv.FormatInt(parsedURL.ProjectID, 10))
					bucketID = strconv.FormatInt(parsedURL.ProjectID, 10)
				}
			}

			client, err := f.ApiClient()
			if err != nil {
				return err
			}
			uploadOps := client.Uploads()

			resolvedProjectID, err := f.ProjectID()
			if err != nil {
				return err
			}
			if bucketID == "" {
				bucketID = resolvedProjectID
			}

			comment, err := client.GetComment(f.Context(), resolvedProjectID, commentID)
			if err != nil {
				return fmt.Errorf("failed to get comment: %w", err)
			}

			sources := []download.AttachmentSource{
				{Label: "comment", Content: comment.Content},
			}

			_, err = download.DownloadFromSources(f.Context(), uploadOps, bucketID, sources, download.Options{
				OutputDir:       outputDir,
				Overwrite:       overwrite,
				AttachmentIndex: attachmentIndex,
			})
			return err
		},
	}

	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")
	cmd.Flags().StringVarP(&outputDir, "output-dir", "o", "", "Directory to save attachments (default: current directory)")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing files without prompting")
	cmd.Flags().IntVar(&attachmentIndex, "attachment", 0, "Download only specified attachment (1-based index)")

	return cmd
}
