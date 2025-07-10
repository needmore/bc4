package ui

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
)

// SimpleTable provides GitHub CLI-style table rendering
type SimpleTable struct {
	writer     io.Writer
	headers    []string
	rows       [][]string
	noColor    bool
	totalWidth int
}

// NewSimpleTable creates a new simple table writer
func NewSimpleTable(w io.Writer, headers []string) *SimpleTable {
	return &SimpleTable{
		writer:  w,
		headers: headers,
		rows:    [][]string{},
		noColor: !IsTerminal(w),
	}
}

// SetMaxWidth sets the maximum width for the table
func (t *SimpleTable) SetMaxWidth(width int) {
	t.totalWidth = width
}

// AddRow adds a row to the table
func (t *SimpleTable) AddRow(row []string) {
	// Simply add the row - let tabwriter handle alignment
	t.rows = append(t.rows, row)
}

// Render outputs the table
func (t *SimpleTable) Render() {
	if len(t.rows) == 0 {
		return
	}

	// For non-TTY output, use simple TSV
	if !IsTerminal(t.writer) {
		t.renderTSV()
		return
	}

	// Use tabwriter for clean column alignment - GitHub CLI style
	tw := tabwriter.NewWriter(t.writer, 0, 0, 3, ' ', 0)
	defer tw.Flush()

	// Render headers if provided and not empty
	if len(t.headers) > 0 {
		hasNonEmptyHeader := false
		for _, h := range t.headers {
			if h != "" {
				hasNonEmptyHeader = true
				break
			}
		}

		if hasNonEmptyHeader {
			// GitHub CLI style: underlined headers (actual underlines, not dashes)
			headerStyle := lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.Color("240"))
			headers := make([]string, len(t.headers))
			
			for i, h := range t.headers {
				if t.noColor {
					headers[i] = h
				} else {
					headers[i] = headerStyle.Render(h)
				}
			}
			
			fmt.Fprintln(tw, strings.Join(headers, "\t")+"\t")
		}
	}

	// Render rows - let tabwriter handle alignment naturally like GitHub CLI
	for _, row := range t.rows {
		// Ensure each row ends with a tab (required for tabwriter)
		fmt.Fprintln(tw, strings.Join(row, "\t")+"\t")
	}
}

// renderTSV outputs in TSV format for non-TTY
func (t *SimpleTable) renderTSV() {
	// Headers
	if len(t.headers) > 0 {
		hasNonEmptyHeader := false
		for _, h := range t.headers {
			if h != "" {
				hasNonEmptyHeader = true
				break
			}
		}
		if hasNonEmptyHeader {
			fmt.Fprintln(t.writer, strings.Join(t.headers, "\t"))
		}
	}

	// Rows
	for _, row := range t.rows {
		// Strip ANSI codes for TSV output
		cleanRow := make([]string, len(row))
		for i, cell := range row {
			cleanRow[i] = stripAnsi(cell)
		}
		fmt.Fprintln(t.writer, strings.Join(cleanRow, "\t"))
	}
}


// truncateWithEllipsis truncates a string to maxLen with ellipsis
func truncateWithEllipsis(s string, maxLen int) string {
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}

	if maxLen < 4 {
		return s[:maxLen]
	}

	// Account for ANSI codes
	cleaned := stripAnsi(s)
	if utf8.RuneCountInString(cleaned) <= maxLen {
		return s
	}

	// Truncate and add ellipsis
	runes := []rune(cleaned)
	truncated := string(runes[:maxLen-3]) + "..."

	return truncated
}

// Helper for status symbols
func StatusSymbol(completed bool, noColor bool) string {
	if completed {
		if noColor {
			return "✓"
		}
		return lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓")
	}
	if noColor {
		return "○"
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("○")
}

// Helper for muted text
func MutedText(text string, noColor bool) string {
	if noColor || text == "" {
		return text
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(text)
}

