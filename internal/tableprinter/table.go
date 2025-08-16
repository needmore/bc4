package tableprinter

import (
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// TablePrinter provides the interface for table rendering, matching GitHub CLI's design
type TablePrinter interface {
	AddHeader([]string, ...fieldOption)
	AddField(string, ...fieldOption)
	EndRow()
	Render() error
}

// fieldOption represents formatting options that can be applied to individual fields
type fieldOption func(*field)

// field represents a single table cell with its formatting options
type field struct {
	Text         string
	TruncateFunc func(int, string) string
	PaddingFunc  func(int, string) string
	ColorFunc    func(string) string
}

// WithTruncate applies a custom truncation function to a field
func WithTruncate(truncateFunc func(int, string) string) fieldOption {
	return func(f *field) {
		f.TruncateFunc = truncateFunc
	}
}

// WithPadding applies a custom padding function to a field
func WithPadding(paddingFunc func(int, string) string) fieldOption {
	return func(f *field) {
		f.PaddingFunc = paddingFunc
	}
}

// WithColor applies a color function to a field
func WithColor(colorFunc func(string) string) fieldOption {
	return func(f *field) {
		f.ColorFunc = colorFunc
	}
}

// New creates a new TablePrinter based on the writer and TTY detection
func New(writer io.Writer, isTTY bool, maxWidth int) TablePrinter {
	if !isTTY {
		return &csvTablePrinter{
			writer: writer,
		}
	}

	return &ttyTablePrinter{
		writer:   writer,
		maxWidth: maxWidth,
		rows:     [][]field{},
	}
}

// IsTTY detects if the writer is a terminal, following GitHub CLI's logic
func IsTTY(w io.Writer) bool {
	// Check for forced TTY mode
	if os.Getenv("BC4_FORCE_TTY") != "" || os.Getenv("FORCE_COLOR") != "" {
		return true
	}

	// Check for forced non-TTY mode
	if os.Getenv("NO_COLOR") != "" || os.Getenv("CLICOLOR") == "0" {
		return false
	}

	if f, ok := w.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}

// GetTerminalWidth returns the terminal width with fallback
func GetTerminalWidth() int {
	defaultWidth := 80
	if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && width > 0 {
		return width
	}
	return defaultWidth
}

// stripAnsi removes ANSI escape sequences for width calculation
func stripAnsi(s string) string {
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

// measureWidth returns the display width of a string, accounting for ANSI codes
func measureWidth(s string) int {
	return len(stripAnsi(s))
}

// defaultTruncate provides the default truncation behavior
func defaultTruncate(maxWidth int, s string) string {
	if measureWidth(s) <= maxWidth {
		return s
	}

	if maxWidth < 4 {
		return s[:maxWidth]
	}

	// Strip ANSI for measurement but preserve in output
	stripped := stripAnsi(s)
	if len(stripped) <= maxWidth-3 {
		return s
	}

	// Truncate the stripped version and add ellipsis
	runes := []rune(stripped)
	return string(runes[:maxWidth-3]) + "..."
}

// defaultPadding provides the default padding behavior
func defaultPadding(width int, s string) string {
	currentWidth := measureWidth(s)
	if currentWidth >= width {
		return s
	}

	padding := strings.Repeat(" ", width-currentWidth)
	return s + padding
}
