package attachments

import (
	"fmt"
	"regexp"
	"strings"
)

// Attachment represents a Basecamp attachment with metadata
type Attachment struct {
	SGID        string
	ContentType string
	Filename    string
	URL         string
	Href        string
	Width       string
	Height      string
	Caption     string
}

// ParseAttachments extracts all bc-attachment elements from HTML content
func ParseAttachments(htmlContent string) []Attachment {
	// Regular expression to match bc-attachment tags with their attributes
	// This handles both self-closing and non-self-closing tags
	// (?s) enables DOTALL mode so . matches newlines
	bcAttachmentPattern := `(?s)<bc-attachment([^>]*)(?:>.*?</bc-attachment>|/>)`
	re := regexp.MustCompile(bcAttachmentPattern)

	matches := re.FindAllStringSubmatch(htmlContent, -1)
	attachments := make([]Attachment, 0, len(matches))

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		attrs := match[1]
		attachment := Attachment{
			SGID:        extractAttribute(attrs, "sgid"),
			ContentType: extractAttribute(attrs, "content-type"),
			Filename:    extractAttribute(attrs, "filename"),
			URL:         extractAttribute(attrs, "url"),
			Href:        extractAttribute(attrs, "href"),
			Width:       extractAttribute(attrs, "width"),
			Height:      extractAttribute(attrs, "height"),
			Caption:     extractAttribute(attrs, "caption"),
		}

		attachments = append(attachments, attachment)
	}

	return attachments
}

// extractAttribute extracts the value of an HTML attribute from a string
func extractAttribute(attrs, attrName string) string {
	// Pattern to match attribute="value" or attribute='value'
	pattern := attrName + `\s*=\s*["']([^"']*)["']`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(attrs)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// IsImage returns true if the attachment is an image based on its content type
func (a *Attachment) IsImage() bool {
	return strings.HasPrefix(a.ContentType, "image/")
}

// GetDisplayName returns the best display name for the attachment
func (a *Attachment) GetDisplayName() string {
	if a.Caption != "" {
		return a.Caption
	}
	if a.Filename != "" {
		return a.Filename
	}
	return "Unnamed attachment"
}

// BuildTag returns a bc-attachment tag for an sgid.
func BuildTag(sgid string) string {
	if sgid == "" {
		return ""
	}
	return fmt.Sprintf(`<bc-attachment sgid="%s"></bc-attachment>`, sgid)
}

// ExtractBucketID extracts the bucket ID from a Basecamp URL
func ExtractBucketID(url string) (string, error) {
	re := regexp.MustCompile(`/buckets/(\d+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1], nil
	}
	return "", fmt.Errorf("could not extract bucket ID from URL: %s", url)
}

// IsBlobURL returns true if the URL is a blob/preview URL that cannot be downloaded via OAuth.
// Blob URLs use Google Cloud Storage with session cookie authentication instead of OAuth tokens.
// These URLs typically contain "/blobs/" or use the "preview.3.basecamp.com" subdomain.
func IsBlobURL(url string) bool {
	return strings.Contains(url, "/blobs/") || strings.Contains(url, "preview.3.basecamp.com")
}

// ExtractUploadIDFromURL extracts the upload ID from a download or app URL
func ExtractUploadIDFromURL(url string) (int64, error) {
	// Pattern for download URL with /uploads/{id}/
	re := regexp.MustCompile(`/uploads/(\d+)(/|$)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		var id int64
		if _, err := fmt.Sscanf(matches[1], "%d", &id); err != nil {
			return 0, fmt.Errorf("failed to parse upload ID: %w", err)
		}
		return id, nil
	}
	return 0, fmt.Errorf("could not extract upload ID from URL: %s", url)
}

// DownloadResult contains the result of attempting to extract download information
type DownloadResult struct {
	UploadID  int64
	SourceURL string // The URL that worked (either URL or Href)
	IsBlobURL bool   // True if this is a blob URL that can't be downloaded via API
	BlobURL   string // The blob URL for user reference (if IsBlobURL is true)
}

// TryExtractUploadID attempts to extract an upload ID from an attachment.
// It first tries the URL field, then falls back to the Href field if available.
// If both fail and the URL is a blob URL, it returns a result indicating the blob URL
// so callers can provide appropriate user guidance.
func TryExtractUploadID(att *Attachment) (*DownloadResult, error) {
	// First try the URL field
	if att.URL != "" {
		if uploadID, err := ExtractUploadIDFromURL(att.URL); err == nil {
			return &DownloadResult{
				UploadID:  uploadID,
				SourceURL: att.URL,
				IsBlobURL: false,
			}, nil
		}
	}

	// Try the Href field as fallback (it might have a different URL format)
	if att.Href != "" && att.Href != att.URL {
		if uploadID, err := ExtractUploadIDFromURL(att.Href); err == nil {
			return &DownloadResult{
				UploadID:  uploadID,
				SourceURL: att.Href,
				IsBlobURL: false,
			}, nil
		}
	}

	// Check if this is a blob URL - these require browser session authentication
	blobURL := ""
	if att.URL != "" && IsBlobURL(att.URL) {
		blobURL = att.URL
	} else if att.Href != "" && IsBlobURL(att.Href) {
		blobURL = att.Href
	}

	if blobURL != "" {
		return &DownloadResult{
			IsBlobURL: true,
			BlobURL:   blobURL,
		}, fmt.Errorf("blob URL requires browser authentication")
	}

	// Neither URL nor Href worked and it's not a blob URL
	return nil, fmt.Errorf("could not extract upload ID from attachment URLs")
}
