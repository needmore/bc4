package attachments

import (
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
