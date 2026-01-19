package todo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/needmore/bc4/internal/attachments"
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
		Use:   "download-attachments [todo-id or URL]",
		Short: "Download attachments from a todo",
		Long: `Download all attachments (images and files) from a todo to local files.

This command fetches attachment metadata from the Basecamp API and downloads
the actual files using OAuth authentication. You can download all attachments
or select specific ones.`,
		Example: `  # Download all attachments from a todo
  bc4 todo download-attachments 123456

  # Download to specific directory
  bc4 todo download-attachments 123456 --output-dir ~/Downloads

  # Download only the first attachment
  bc4 todo download-attachments 123456 --attachment 1

  # Overwrite existing files
  bc4 todo download-attachments 123456 --overwrite`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if accountID != "" {
				f = f.WithAccount(accountID)
			}
			if projectID != "" {
				f = f.WithProject(projectID)
			}

			todoID, parsedURL, err := parser.ParseArgument(args[0])
			if err != nil {
				return fmt.Errorf("invalid todo ID or URL: %s", args[0])
			}

			var bucketID string
			if parsedURL != nil {
				if parsedURL.ResourceType != parser.ResourceTypeTodo {
					return fmt.Errorf("URL is not for a todo: %s", args[0])
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
			todoOps := client.Todos()
			uploadOps := client.Uploads()

			resolvedProjectID, err := f.ProjectID()
			if err != nil {
				return err
			}
			if bucketID == "" {
				bucketID = resolvedProjectID
			}

			todo, err := todoOps.GetTodo(f.Context(), resolvedProjectID, todoID)
			if err != nil {
				return fmt.Errorf("failed to get todo: %w", err)
			}

			atts := attachments.ParseAttachments(todo.Description)
			if len(atts) == 0 {
				fmt.Println("No attachments found in this todo")
				return nil
			}

			if attachmentIndex > 0 {
				if attachmentIndex > len(atts) {
					return fmt.Errorf("attachment index %d out of range (todo has %d attachments)", attachmentIndex, len(atts))
				}
				atts = []attachments.Attachment{atts[attachmentIndex-1]}
			}

			if outputDir == "" {
				outputDir = "."
			}

			successful := 0
			failed := 0
			ctx := context.Background()

			for i, att := range atts {
				displayIndex := i + 1
				if attachmentIndex > 0 {
					displayIndex = attachmentIndex
				}

				fmt.Printf("Downloading attachment %d/%d: %s\n", displayIndex, len(atts), att.GetDisplayName())

				uploadID, err := attachments.ExtractUploadIDFromURL(att.URL)
				if err != nil {
					fmt.Printf("  ✗ Failed: %v\n", err)
					failed++
					continue
				}

				upload, err := uploadOps.GetUpload(ctx, bucketID, uploadID)
				if err != nil {
					fmt.Printf("  ✗ Failed to get upload details: %v\n", err)
					failed++
					continue
				}

				filename := upload.Filename
				destPath := filepath.Join(outputDir, filename)

				if !overwrite {
					if _, err := os.Stat(destPath); err == nil {
						fmt.Printf("  ⚠ File already exists: %s (use --overwrite to replace)\n", destPath)
						fmt.Println("  Skipping...")
						continue
					}
				}

				err = uploadOps.DownloadAttachment(ctx, upload.DownloadURL, destPath)
				if err != nil {
					fmt.Printf("  ✗ Failed to download: %v\n", err)
					failed++
					continue
				}

				sizeStr := formatByteSize(upload.ByteSize)
				fmt.Printf("  ✓ Downloaded: %s (%s)\n", destPath, sizeStr)
				successful++
			}

			fmt.Println()
			if successful > 0 {
				fmt.Printf("Successfully downloaded: %d/%d attachments\n", successful, len(atts))
			}
			if failed > 0 {
				fmt.Printf("Failed: %d attachments\n", failed)
				return fmt.Errorf("some attachments failed to download")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")
	cmd.Flags().StringVarP(&outputDir, "output-dir", "o", "", "Directory to save attachments (default: current directory)")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing files without prompting")
	cmd.Flags().IntVar(&attachmentIndex, "attachment", 0, "Download only specified attachment (1-based index)")

	return cmd
}

func formatByteSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
