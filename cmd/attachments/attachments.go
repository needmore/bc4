package attachments

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/needmore/bc4/internal/attachments"
	"github.com/needmore/bc4/internal/ui/tableprinter"
)

// DisplayAttachments displays a list of attachments in a formatted table
func DisplayAttachments(htmlContent string) error {
	atts := attachments.ParseAttachments(htmlContent)

	if len(atts) == 0 {
		fmt.Println("No attachments found.")
		return nil
	}

	// Create table for displaying attachments
	table := tableprinter.New(os.Stdout)

	// Add headers
	table.AddHeader("#", "FILENAME", "TYPE", "SIZE", "URL")

	// Add each attachment
	for i, att := range atts {
		// Row number
		table.AddField(fmt.Sprintf("%d", i+1))

		// Filename/Caption
		table.AddField(att.GetDisplayName())

		// Content type
		contentType := att.ContentType
		if contentType == "" {
			contentType = "unknown"
		}
		table.AddField(contentType)

		// Size (if image, show dimensions)
		size := "-"
		if att.IsImage() && att.Width != "" && att.Height != "" {
			size = fmt.Sprintf("%s×%s", att.Width, att.Height)
		}
		table.AddField(size)

		// URL
		url := att.URL
		if url == "" {
			url = att.Href
		}
		if url == "" {
			url = "-"
		}
		table.AddField(url)

		table.EndRow()
	}

	return table.Render()
}

// DisplayAttachmentsWithStyle displays attachments with styled output (for view commands)
func DisplayAttachmentsWithStyle(htmlContent string) string {
	atts := attachments.ParseAttachments(htmlContent)

	if len(atts) == 0 {
		return ""
	}

	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("8"))
	result := fmt.Sprintf("\n%s\n", labelStyle.Render("Attachments:"))

	for i, att := range atts {
		displayName := att.GetDisplayName()

		// Show type indicator
		typeIndicator := "📎"
		if att.IsImage() {
			typeIndicator = "🖼️"
		}

		result += fmt.Sprintf("  %d. %s %s", i+1, typeIndicator, displayName)

		// Add dimensions for images
		if att.IsImage() && att.Width != "" && att.Height != "" {
			result += fmt.Sprintf(" (%s×%s)", att.Width, att.Height)
		}

		result += "\n"

		// Add URL if available
		url := att.URL
		if url == "" {
			url = att.Href
		}
		if url != "" {
			result += fmt.Sprintf("     URL: %s\n", url)
		}
	}

	return result
}
