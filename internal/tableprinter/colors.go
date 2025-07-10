package tableprinter

import (
	"fmt"
	"os"
)

// ColorScheme provides state-based color functions following GitHub CLI's design
type ColorScheme struct {
	// State colors for different entity states
	Green   func(string) string // Active/Open items
	Red     func(string) string // Completed/Closed items
	Magenta func(string) string // Special/Merged items
	Gray    func(string) string // Draft/Inactive items
	Cyan    func(string) string // Names/Identifiers
	Yellow  func(string) string // Private/Warning items
	Muted   func(string) string // Secondary info/timestamps
	Bold    func(string) string // Emphasis
}

// NewColorScheme creates a new color scheme with TTY detection
func NewColorScheme() *ColorScheme {
	if !shouldUseColor() {
		return &ColorScheme{
			Green:   noColor,
			Red:     noColor,
			Magenta: noColor,
			Gray:    noColor,
			Cyan:    noColor,
			Yellow:  noColor,
			Muted:   noColor,
			Bold:    noColor,
		}
	}

	return &ColorScheme{
		Green:   ansiColor(32), // Green
		Red:     ansiColor(31), // Red
		Magenta: ansiColor(35), // Magenta
		Gray:    ansiColor(90), // Bright black (gray)
		Cyan:    ansiColor(36), // Cyan
		Yellow:  ansiColor(33), // Yellow
		Muted:   ansiColor(37), // Light gray
		Bold:    ansiBold,      // Bold formatting
	}
}

// shouldUseColor determines if color output should be used
func shouldUseColor() bool {
	// Check for forced color
	if os.Getenv("BC4_FORCE_TTY") != "" || os.Getenv("FORCE_COLOR") != "" {
		return true
	}

	// Check for disabled color
	if os.Getenv("NO_COLOR") != "" || os.Getenv("CLICOLOR") == "0" {
		return false
	}

	// Default to TTY detection
	return IsTTY(os.Stdout)
}

// ColorFromString returns a color function based on a state string
// This matches GitHub CLI's pattern for dynamic color selection
func (cs *ColorScheme) ColorFromString(state string) func(string) string {
	switch state {
	case "open", "active", "incomplete":
		return cs.Green
	case "closed", "completed", "done":
		return cs.Red
	case "merged", "special":
		return cs.Magenta
	case "draft", "archived", "inactive":
		return cs.Gray
	case "private", "warning":
		return cs.Yellow
	default:
		return noColor
	}
}

// noColor returns the input string without any color formatting
func noColor(s string) string {
	return s
}

// ansiColor returns a function that wraps text in ANSI color codes
func ansiColor(code int) func(string) string {
	return func(s string) string {
		return fmt.Sprintf("\x1b[%dm%s\x1b[0m", code, s)
	}
}

// ansiBold returns a function that wraps text in ANSI bold formatting
func ansiBold(s string) string {
	return fmt.Sprintf("\x1b[1m%s\x1b[0m", s)
}

// Predefined color functions for common Basecamp entity states
func (cs *ColorScheme) ProjectActive(s string) string   { return cs.Green(s) }
func (cs *ColorScheme) ProjectArchived(s string) string { return cs.Gray(s) }
func (cs *ColorScheme) TodoIncomplete(s string) string  { return cs.Green(s) }
func (cs *ColorScheme) TodoCompleted(s string) string   { return cs.Red(s) }
func (cs *ColorScheme) TodoListName(s string) string    { return cs.Cyan(s) }
func (cs *ColorScheme) AccountName(s string) string     { return cs.Cyan(s) }
func (cs *ColorScheme) Timestamp(s string) string       { return cs.Muted(s) }
func (cs *ColorScheme) Description(s string) string     { return cs.Muted(s) }
func (cs *ColorScheme) ID(s string) string              { return cs.Bold(s) }
