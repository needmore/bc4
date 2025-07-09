package ui

import (
	"os"
	"strings"

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
	s.Selected = DefaultTableSelectedStyle() // Use subtle highlighting
	t.SetStyles(s)
	return t
}

// StyleTableForList applies styling for list views with subtle default highlighting
func StyleTableForList(t table.Model) table.Model {
	s := table.DefaultStyles()
	s.Header = lipgloss.NewStyle().Height(0)                           // Hide header completely
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

// ColumnWidthConfig defines how to calculate column widths
type ColumnWidthConfig struct {
	// MinWidths defines minimum width for each column
	MinWidths []int
	// MaxWidths defines maximum width for each column (0 = no limit)
	MaxWidths []int
	// PreferredWidths defines preferred width for each column
	PreferredWidths []int
	// FlexColumns marks which columns should expand to fill available space
	FlexColumns []bool
}

// CalculateColumnWidths calculates optimal column widths based on terminal width
// This is inspired by GitHub CLI's approach to responsive table layouts
func CalculateColumnWidths(termWidth int, config ColumnWidthConfig) []int {
	numCols := len(config.MinWidths)
	if numCols == 0 {
		return []int{}
	}

	// Start with preferred widths
	widths := make([]int, numCols)
	copy(widths, config.PreferredWidths)

	// Ensure minimum widths
	for i := 0; i < numCols; i++ {
		if i < len(config.MinWidths) && widths[i] < config.MinWidths[i] {
			widths[i] = config.MinWidths[i]
		}
	}

	// Apply maximum widths
	for i := 0; i < numCols; i++ {
		if i < len(config.MaxWidths) && config.MaxWidths[i] > 0 && widths[i] > config.MaxWidths[i] {
			widths[i] = config.MaxWidths[i]
		}
	}

	// Calculate total width including separators (assume 3 chars per separator)
	separatorWidth := (numCols - 1) * 3
	totalWidth := separatorWidth
	for _, w := range widths {
		totalWidth += w
	}

	// If we're within terminal width, we're done
	if totalWidth <= termWidth {
		// If we have extra space, distribute to flex columns
		extraSpace := termWidth - totalWidth
		flexCount := 0
		for i, flex := range config.FlexColumns {
			if i < numCols && flex {
				flexCount++
			}
		}

		if flexCount > 0 && extraSpace > 0 {
			perColumn := extraSpace / flexCount
			remainder := extraSpace % flexCount
			for i, flex := range config.FlexColumns {
				if i < numCols && flex {
					widths[i] += perColumn
					if remainder > 0 {
						widths[i]++
						remainder--
					}
					// Apply max width constraint
					if i < len(config.MaxWidths) && config.MaxWidths[i] > 0 && widths[i] > config.MaxWidths[i] {
						widths[i] = config.MaxWidths[i]
					}
				}
			}
		}
		return widths
	}

	// We need to shrink columns
	availableWidth := termWidth - separatorWidth
	if availableWidth < numCols*5 { // Too narrow, just use minimum widths
		for i := 0; i < numCols; i++ {
			if i < len(config.MinWidths) {
				widths[i] = config.MinWidths[i]
			} else {
				widths[i] = 5
			}
		}
		return widths
	}

	// Proportionally shrink columns
	shrinkFactor := float64(availableWidth) / float64(totalWidth-separatorWidth)
	for i := 0; i < numCols; i++ {
		newWidth := int(float64(widths[i]) * shrinkFactor)
		// Ensure minimum width
		if i < len(config.MinWidths) && newWidth < config.MinWidths[i] {
			newWidth = config.MinWidths[i]
		}
		widths[i] = newWidth
	}

	return widths
}

// MeasureStringWidth returns the display width of a string, accounting for ANSI codes
func MeasureStringWidth(s string) int {
	// Strip ANSI codes for width calculation
	stripped := stripAnsi(s)
	return len(stripped)
}

// stripAnsi removes ANSI escape sequences from a string
func stripAnsi(s string) string {
	// Simple ANSI stripping - in production, use a proper library
	var result strings.Builder
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
		} else if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
