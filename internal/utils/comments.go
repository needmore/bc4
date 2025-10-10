package utils

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/needmore/bc4/internal/api"
)

// FormatCommentsForDisplay formats a list of comments for display in a pager
func FormatCommentsForDisplay(comments []api.Comment) (string, error) {
	if len(comments) == 0 {
		return "", nil
	}

	var buf bytes.Buffer

	// Title for comments section
	fmt.Fprintf(&buf, "\n%s\n", strings.Repeat("=", 50))
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	fmt.Fprintf(&buf, "%s\n", titleStyle.Render(fmt.Sprintf("Comments (%d)", len(comments))))
	fmt.Fprintf(&buf, "%s\n\n", strings.Repeat("=", 50))

	// Render each comment
	for i, comment := range comments {
		if i > 0 {
			fmt.Fprintf(&buf, "\n%s\n\n", strings.Repeat("-", 50))
		}

		// Comment header with author and date
		metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		fmt.Fprintf(&buf, "%s\n\n", metaStyle.Render(fmt.Sprintf("Comment #%d by %s • %s",
			comment.ID,
			comment.Creator.Name,
			comment.CreatedAt.Format("Jan 2, 2006 at 3:04 PM"))))

		// Render content with glamour
		r, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(80),
		)
		if err != nil {
			// Fallback to plain text if glamour fails
			fmt.Fprintf(&buf, "%s\n", comment.Content)
			continue
		}

		rendered, err := r.Render(comment.Content)
		if err != nil {
			// Fallback to plain text if rendering fails
			fmt.Fprintf(&buf, "%s\n", comment.Content)
			continue
		}

		fmt.Fprint(&buf, rendered)
	}

	return buf.String(), nil
}
