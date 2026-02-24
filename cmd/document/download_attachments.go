package document

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
	var includeComments bool

	cmd := &cobra.Command{
		Use:   "download-attachments [document-id or URL]",
		Short: "Download attachments from a document",
		Long: `Download all attachments (images and files) from a document to local files.

This command fetches attachment metadata from the Basecamp API and downloads
the actual files using OAuth authentication. You can download all attachments
or select specific ones.

You can specify the document using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/documents/12345")`,
		Example: `  # Download all attachments from a document
  bc4 document download-attachments 123456

  # Download to specific directory
  bc4 document download-attachments 123456 --output-dir ~/Downloads

  # Download only the first attachment
  bc4 document download-attachments 123456 --attachment 1

  # Overwrite existing files
  bc4 document download-attachments 123456 --overwrite

  # Include attachments from comments
  bc4 document download-attachments 123456 --include-comments

  # Using Basecamp URL
  bc4 document download-attachments https://3.basecamp.com/123/buckets/456/documents/789`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if accountID != "" {
				f = f.WithAccount(accountID)
			}
			if projectID != "" {
				f = f.WithProject(projectID)
			}

			documentID, parsedURL, err := parser.ParseArgument(args[0])
			if err != nil {
				return fmt.Errorf("invalid document ID or URL: %s", args[0])
			}

			var bucketID string
			if parsedURL != nil {
				if parsedURL.ResourceType != parser.ResourceTypeDocument {
					return fmt.Errorf("URL is not for a document: %s", args[0])
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

			document, err := client.GetDocument(f.Context(), resolvedProjectID, documentID)
			if err != nil {
				return fmt.Errorf("failed to get document: %w", err)
			}

			sources := []download.AttachmentSource{
				{Label: "document", Content: document.Content},
			}

			if includeComments {
				comments, err := client.ListComments(f.Context(), resolvedProjectID, document.ID)
				if err != nil {
					return fmt.Errorf("failed to fetch comments: %w", err)
				}
				for _, c := range comments {
					sources = append(sources, download.AttachmentSource{
						Label:   fmt.Sprintf("comment #%d by %s", c.ID, c.Creator.Name),
						Content: c.Content,
					})
				}
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
	cmd.Flags().BoolVar(&includeComments, "include-comments", false, "Also download attachments from comments on this document")

	return cmd
}
