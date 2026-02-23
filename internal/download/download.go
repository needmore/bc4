package download

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/attachments"
)

// AttachmentSource describes where attachments came from (for display).
type AttachmentSource struct {
	Label   string // e.g. "card", "comment #12345 by Alice"
	Content string // HTML content containing <bc-attachment> tags
}

// Options configures the download behavior.
type Options struct {
	OutputDir       string
	Overwrite       bool
	AttachmentIndex int // 1-based; 0 means "all"
}

// Result tracks the outcome of a download run.
type Result struct {
	Successful int
	Failed     int
	Skipped    int
	Total      int
}

// DownloadFromSources parses attachments from one or more HTML content sources
// and downloads them. When AttachmentIndex is set, it applies to the combined
// attachment list across all sources.
func DownloadFromSources(
	ctx context.Context,
	uploadOps api.UploadOperations,
	bucketID string,
	sources []AttachmentSource,
	opts Options,
) (*Result, error) {
	type taggedAttachment struct {
		att    attachments.Attachment
		source string
	}

	var allAtts []taggedAttachment
	for _, src := range sources {
		parsed := attachments.ParseAttachments(src.Content)
		for _, att := range parsed {
			allAtts = append(allAtts, taggedAttachment{att: att, source: src.Label})
		}
	}

	if len(allAtts) == 0 {
		fmt.Println("No attachments found")
		return &Result{}, nil
	}

	originalCount := len(allAtts)

	// Filter to specific attachment if requested
	if opts.AttachmentIndex > 0 {
		if opts.AttachmentIndex > originalCount {
			return nil, fmt.Errorf("attachment index %d out of range (found %d attachments)", opts.AttachmentIndex, originalCount)
		}
		allAtts = []taggedAttachment{allAtts[opts.AttachmentIndex-1]}
	}

	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = "."
	}

	multiSource := len(sources) > 1
	result := &Result{Total: len(allAtts)}

	for i, ta := range allAtts {
		displayIndex := i + 1
		if opts.AttachmentIndex > 0 {
			displayIndex = opts.AttachmentIndex
		}

		// Show source label when downloading from multiple sources
		sourcePrefix := ""
		if multiSource {
			sourcePrefix = fmt.Sprintf("[%s] ", ta.source)
		}

		if opts.AttachmentIndex > 0 {
			fmt.Printf("%sDownloading attachment %d: %s\n", sourcePrefix, displayIndex, ta.att.GetDisplayName())
		} else {
			fmt.Printf("%sDownloading attachment %d/%d: %s\n", sourcePrefix, displayIndex, originalCount, ta.att.GetDisplayName())
		}

		// Try to extract upload ID from URL or Href
		extractResult, err := attachments.TryExtractUploadID(&ta.att)
		if err != nil {
			if extractResult != nil && extractResult.IsBlobURL {
				fmt.Println("  ✗ Cannot download via API: This attachment uses a browser-only URL")
				fmt.Printf("    URL: %s\n", extractResult.BlobURL)
				fmt.Println("    Tip: Open this URL in your browser while logged into Basecamp to download")
			} else {
				fmt.Printf("  ✗ Failed: %v\n", err)
			}
			result.Failed++
			continue
		}

		// Get full upload details including download URL
		upload, err := uploadOps.GetUpload(ctx, bucketID, extractResult.UploadID)
		if err != nil {
			fmt.Printf("  ✗ Failed to get upload details: %v\n", err)
			result.Failed++
			continue
		}

		// Sanitize filename for filesystem safety
		filename := SanitizeFilename(upload.Filename)
		destPath := filepath.Join(outputDir, filename)

		// Check if file exists
		if !opts.Overwrite {
			if _, err := os.Stat(destPath); err == nil {
				fmt.Printf("  ⚠ File already exists: %s (use --overwrite to replace)\n", destPath)
				fmt.Println("  Skipping...")
				result.Skipped++
				continue
			}
		}

		// Download the attachment
		err = uploadOps.DownloadAttachment(ctx, upload.DownloadURL, destPath)
		if err != nil {
			fmt.Printf("  ✗ Failed to download: %v\n", err)
			result.Failed++
			continue
		}

		sizeStr := FormatByteSize(upload.ByteSize)
		fmt.Printf("  ✓ Downloaded: %s (%s)\n", destPath, sizeStr)
		result.Successful++
	}

	// Print summary
	fmt.Println()
	if result.Successful > 0 {
		fmt.Printf("Successfully downloaded: %d/%d attachments\n", result.Successful, result.Total)
	}
	if result.Failed > 0 {
		fmt.Printf("Failed: %d attachments\n", result.Failed)
		return result, fmt.Errorf("some attachments failed to download")
	}

	return result, nil
}

// SanitizeFilename removes or replaces characters that are unsafe for filenames
// to prevent path traversal attacks and filesystem errors.
func SanitizeFilename(filename string) string {
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

	if cleaned == "" || cleaned == "." || cleaned == ".." {
		cleaned = "attachment"
	}

	return cleaned
}

// FormatByteSize formats a byte size in a human-readable format.
func FormatByteSize(bytes int64) string {
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
