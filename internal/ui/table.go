package ui

import (
	"os"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// GetTerminalWidth returns the current terminal width with a fallback
func GetTerminalWidth() int {
	defaultWidth := 100
	if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && width > 0 {
		return width - 4 // Leave some margin for borders
	}
	return defaultWidth
}

// GetTerminalHeight returns the current terminal height with a fallback
func GetTerminalHeight() int {
	defaultHeight := 24
	if _, height, err := term.GetSize(int(os.Stdout.Fd())); err == nil && height > 0 {
		return height
	}
	return defaultHeight
}

// CalculateTableHeight calculates appropriate table height based on terminal and row count
func CalculateTableHeight(terminalHeight, rowCount int) int {
	// Leave room for borders, header, and help text
	tableHeight := terminalHeight - 6
	if tableHeight < 10 {
		tableHeight = 10 // Minimum height
	}
	// Don't make it taller than needed
	if tableHeight > rowCount+2 {
		tableHeight = rowCount + 2
	}
	return tableHeight
}

// StyleTable applies common styling to a table (for interactive select views)
func StyleTable(t table.Model) table.Model {
	s := table.DefaultStyles()
	s.Header = lipgloss.NewStyle().Height(0) // Hide header
	s.Selected = DefaultTableSelectedStyle()
	t.SetStyles(s)
	return t
}

// StyleTableForDisplay applies styling for non-interactive display tables
func StyleTableForDisplay(t table.Model) table.Model {
	s := table.DefaultStyles()
	s.Header = lipgloss.NewStyle().Height(0) // Hide header completely
	s.Selected = DefaultTableSelectedStyle()  // Use subtle highlighting
	t.SetStyles(s)
	return t
}

// StyleTableForList applies styling for list views with subtle default highlighting
func StyleTableForList(t table.Model) table.Model {
	s := table.DefaultStyles()
	s.Header = lipgloss.NewStyle().Height(0) // Hide header completely
	s.Selected = lipgloss.NewStyle().Background(lipgloss.Color("229")) // Subtle yellow
	t.SetStyles(s)
	return t
}

// TruncateString truncates a string to fit within maxLen, adding "..." if needed
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 4 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// HighlightTableRow applies the default row highlight style to all cells in a row
func HighlightTableRow(cells []string, widths []int) []string {
	highlighted := make([]string, len(cells))
	for i, cell := range cells {
		// Pad cell to column width and apply style
		if i < len(widths) {
			highlighted[i] = DefaultRowStyle.Width(widths[i]).Render(cell)
		} else {
			highlighted[i] = DefaultRowStyle.Render(cell)
		}
	}
	return highlighted
}