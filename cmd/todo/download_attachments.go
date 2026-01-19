package todo

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
or select specific ones.

You can specify the card using either:
- A numeric ID (e.g., "12345")
- A Basecamp URL (e.g., "https://3.basecamp.com/1234567/buckets/89012345/todos/12345")`,
		Example: `  # Download all attachments from a todo
  bc4 todo download-attachments 123456

  # Download to specific directory
  bc4 todo download-attachments 123456 --output-dir ~/Downloads

  # Download only the first attachment
  bc4 todo download-attachments 123456 --attachment 1

  # Overwrite existing files
  bc4 todo download-attachments 123456 --overwrite

  # Using Basecamp URL
  bc4 todo download-attachments https://3.basecamp.com/123/buckets/456/todos/789`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Apply account override if specified
			if accountID != "" {
				f = f.WithAccount(accountID)
			}

			// Apply project override if specified
			if projectID != "" {
				f = f.WithProject(projectID)
			}

			// Parse card ID (could be numeric ID or URL)
			cardID, parsedURL, err := parser.ParseArgument(args[0])
			if err != nil {
				return fmt.Errorf("invalid card ID or URL: %s", args[0])
			}

			// If a URL was parsed, override account and project IDs if provided
			var bucketID string
			if parsedURL != nil {
				if parsedURL.ResourceType != parser.ResourceTypeTodo {
					return fmt.Errorf("URL is not for a card: %s", args[0])
				}
				if parsedURL.AccountID > 0 {
					f = f.WithAccount(strconv.FormatInt(parsedURL.AccountID, 10))
				}
				if parsedURL.ProjectID > 0 {
					f = f.WithProject(strconv.FormatInt(parsedURL.ProjectID, 10))
					bucketID = strconv.FormatInt(parsedURL.ProjectID, 10)
				}
			}

			// Get API client from factory
			client, err := f.ApiClient()
			if err != nil {
				return err
			}
			todoOps := client.Todos()
			uploadOps := client.Uploads()

			// Get resolved project ID (bucket ID)
			resolvedProjectID, err := f.ProjectID()
			if err != nil {
				return err
			}
			if bucketID == "" {
				bucketID = resolvedProjectID
			}

			// Fetch the card details
			todo, err := todoOps.GetTodo(f.Context(), resolvedProjectID, cardID)
			if err != nil {
				return fmt.Errorf("failed to get card: %w", err)
			}

			// Parse attachments from card content
			atts := attachments.ParseAttachments(todo.Description)
			if len(atts) == 0 {
				fmt.Println("No attachments found in this card")
				return nil
			}

			// Store original count before filtering
			originalCount := len(atts)

			// Filter to specific attachment if requested
			if attachmentIndex > 0 {
				if attachmentIndex > originalCount {
					return fmt.Errorf("attachment index %d out of range (todo has %d attachments)", attachmentIndex, originalCount)
				}
				atts = []attachments.Attachment{atts[attachmentIndex-1]}
			}

			// Use current directory if no output directory specified
			if outputDir == "" {
				outputDir = "."
			}

			// Download each attachment
			successful := 0
			failed := 0
			ctx := f.Context()

			for i, att := range atts {
				displayIndex := i + 1
				if attachmentIndex > 0 {
					displayIndex = attachmentIndex
				}

				// Show appropriate progress message based on whether filtering
				if attachmentIndex > 0 {
					fmt.Printf("Downloading attachment %d: %s\n", displayIndex, att.GetDisplayName())
				} else {
					fmt.Printf("Downloading attachment %d/%d: %s\n", displayIndex, originalCount, att.GetDisplayName())
				}

				// Extract upload ID from the URL
				uploadID, err := attachments.ExtractUploadIDFromURL(att.URL)
				if err != nil {
					fmt.Printf("  ✗ Failed: %v\n", err)
					failed++
					continue
				}

				// Get full upload details including download URL
				upload, err := uploadOps.GetUpload(ctx, bucketID, uploadID)
				if err != nil {
					fmt.Printf("  ✗ Failed to get upload details: %v\n", err)
					failed++
					continue
				}

				// Sanitize filename for filesystem safety
				filename := sanitizeFilename(upload.Filename)
				destPath := filepath.Join(outputDir, filename)

				// Check if file exists
				if !overwrite {
					if _, err := os.Stat(destPath); err == nil {
						fmt.Printf("  ⚠ File already exists: %s (use --overwrite to replace)\n", destPath)
						fmt.Println("  Skipping...")
						continue
					}
				}

				// Download the attachment
				err = uploadOps.DownloadAttachment(ctx, upload.DownloadURL, destPath)
				if err != nil {
					fmt.Printf("  ✗ Failed to download: %v\n", err)
					failed++
					continue
				}

				// Format file size
				sizeStr := formatByteSize(upload.ByteSize)
				fmt.Printf("  ✓ Downloaded: %s (%s)\n", destPath, sizeStr)
				successful++
			}

			// Print summary
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

	// Add flags
	cmd.Flags().StringVarP(&accountID, "account", "a", "", "Specify account ID")
	cmd.Flags().StringVarP(&projectID, "project", "p", "", "Specify project ID")
	cmd.Flags().StringVarP(&outputDir, "output-dir", "o", "", "Directory to save attachments (default: current directory)")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing files without prompting")
	cmd.Flags().IntVar(&attachmentIndex, "attachment", 0, "Download only specified attachment (1-based index)")

	return cmd
}

// sanitizeFilename removes or replaces characters that are unsafe for filenames
// to prevent path traversal attacks and filesystem errors
func sanitizeFilename(filename string) string {
	// Remove path separators to prevent directory traversal
	cleaned := filepath.Base(filename)
	
	// Remove null bytes and other control characters
	cleaned = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, cleaned)
	
	// Replace filesystem-unsafe characters with underscores
	unsafe := []string{"<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range unsafe {
		cleaned = strings.ReplaceAll(cleaned, char, "_")
	}
	
	// Prevent empty filenames
	if cleaned == "" || cleaned == "." || cleaned == ".." {
		cleaned = "attachment"
	}
	
	return cleaned
}

// formatByteSize formats a byte size in a human-readable format
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
