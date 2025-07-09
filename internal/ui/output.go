package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"golang.org/x/term"
)

// OutputFormat represents the desired output format
type OutputFormat string

const (
	// OutputFormatTable renders as a human-readable table
	OutputFormatTable OutputFormat = "table"
	// OutputFormatJSON renders as JSON
	OutputFormatJSON OutputFormat = "json"
	// OutputFormatTSV renders as tab-separated values
	OutputFormatTSV OutputFormat = "tsv"
)

// ParseOutputFormat parses a string into an OutputFormat
func ParseOutputFormat(s string) (OutputFormat, error) {
	switch strings.ToLower(s) {
	case "table", "":
		return OutputFormatTable, nil
	case "json":
		return OutputFormatJSON, nil
	case "tsv":
		return OutputFormatTSV, nil
	default:
		return "", fmt.Errorf("unknown output format: %s", s)
	}
}

// IsTerminal returns true if the given writer is a terminal
func IsTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		// Use the same check as in table.go
		fd := int(f.Fd())
		return isatty(fd)
	}
	return false
}

// isatty checks if the given file descriptor is a terminal
func isatty(fd int) bool {
	// Check if we're forcing TTY mode
	if os.Getenv("BC4_FORCE_TTY") != "" || os.Getenv("FORCE_COLOR") != "" {
		return true
	}
	// Try to get terminal size - if it works, it's a terminal
	_, _, err := term.GetSize(fd)
	return err == nil
}

// OutputConfig holds configuration for output formatting
type OutputConfig struct {
	Format    OutputFormat
	NoHeaders bool
	NoColor   bool
	MaxWidth  int
	Writer    io.Writer
}

// NewOutputConfig creates a new output configuration with defaults
func NewOutputConfig(w io.Writer) *OutputConfig {
	// Check for NO_COLOR environment variable
	noColor := os.Getenv("NO_COLOR") != ""

	// Check for CLICOLOR environment variable
	if os.Getenv("CLICOLOR") == "0" {
		noColor = true
	}

	// Force color if requested
	if os.Getenv("BC4_FORCE_TTY") != "" || os.Getenv("FORCE_COLOR") != "" {
		noColor = false
	}

	return &OutputConfig{
		Format:    OutputFormatTable,
		NoHeaders: false,
		NoColor:   noColor || !IsTerminal(w),
		MaxWidth:  GetTerminalWidth(),
		Writer:    w,
	}
}

// TableWriter provides an interface for writing tabular data
type TableWriter interface {
	AddHeader(columns []string)
	AddRow(values []string)
	Render() error
}

// simpleTableWriter writes simple tab-separated values
type simpleTableWriter struct {
	config  *OutputConfig
	headers []string
	rows    [][]string
}

// NewTableWriter creates a new table writer based on the output config
func NewTableWriter(config *OutputConfig) TableWriter {
	return &simpleTableWriter{
		config: config,
		rows:   [][]string{},
	}
}

func (w *simpleTableWriter) AddHeader(columns []string) {
	w.headers = columns
}

func (w *simpleTableWriter) AddRow(values []string) {
	w.rows = append(w.rows, values)
}

func (w *simpleTableWriter) Render() error {
	switch w.config.Format {
	case OutputFormatJSON:
		return w.renderJSON()
	case OutputFormatTSV:
		return w.renderTSV()
	default:
		return w.renderTable()
	}
}

func (w *simpleTableWriter) renderJSON() error {
	// Convert rows to array of objects
	data := []map[string]string{}
	for _, row := range w.rows {
		obj := make(map[string]string)
		for i, col := range row {
			if i < len(w.headers) {
				obj[w.headers[i]] = col
			}
		}
		data = append(data, obj)
	}

	encoder := json.NewEncoder(w.config.Writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func (w *simpleTableWriter) renderTSV() error {
	// Write headers if not disabled
	if !w.config.NoHeaders && len(w.headers) > 0 {
		fmt.Fprintln(w.config.Writer, strings.Join(w.headers, "\t"))
	}

	// Write rows
	for _, row := range w.rows {
		fmt.Fprintln(w.config.Writer, strings.Join(row, "\t"))
	}

	return nil
}

func (w *simpleTableWriter) renderTable() error {
	// For non-TTY output, use TSV
	if !IsTerminal(w.config.Writer) {
		return w.renderTSV()
	}

	// Use lipgloss table for pretty rendering
	if len(w.rows) == 0 {
		return nil
	}

	// Calculate column widths based on content
	colCount := len(w.headers)
	if colCount == 0 && len(w.rows) > 0 {
		colCount = len(w.rows[0])
	}

	// Measure content to determine column widths
	maxWidths := make([]int, colCount)

	// Consider headers
	for i, h := range w.headers {
		if len(h) > maxWidths[i] {
			maxWidths[i] = len(h)
		}
	}

	// Consider rows
	for _, row := range w.rows {
		for i, cell := range row {
			if i < colCount && len(cell) > maxWidths[i] {
				maxWidths[i] = len(cell)
			}
		}
	}

	// Apply constraints based on terminal width
	config := ColumnWidthConfig{
		MinWidths:       make([]int, colCount),
		MaxWidths:       make([]int, colCount),
		PreferredWidths: maxWidths,
		FlexColumns:     make([]bool, colCount),
	}

	// Set reasonable minimums and maximums
	for i := range config.MinWidths {
		config.MinWidths[i] = 8
		// Set sensible max widths for all but last column
		if i < colCount-1 {
			config.MaxWidths[i] = 40
		}
	}

	// Last column is flexible
	if colCount > 0 {
		config.FlexColumns[colCount-1] = true
	}

	// Calculate optimal widths
	widths := CalculateColumnWidths(w.config.MaxWidth, config)

	// Create lipgloss table with minimal styling
	t := table.New().Width(w.config.MaxWidth)

	// Only add borders in color mode
	if !w.config.NoColor {
		t = t.Border(lipgloss.NormalBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240")))
	} else {
		// Use a simpler border for no-color mode
		t = t.Border(lipgloss.RoundedBorder())
	}

	// Add headers if provided
	if !w.config.NoHeaders && len(w.headers) > 0 {
		t = t.Headers(w.headers...)
	}

	// Apply width constraints to each column
	styledRows := make([][]string, len(w.rows))
	for i, row := range w.rows {
		styledRows[i] = make([]string, len(row))
		for j, cell := range row {
			if j < len(widths) {
				// Truncate content if needed
				styledRows[i][j] = TruncateString(cell, widths[j])
			} else {
				styledRows[i][j] = cell
			}
		}
		t = t.Row(styledRows[i]...)
	}

	// Style function for headers (only if color is enabled)
	if !w.config.NoColor && !w.config.NoHeaders && len(w.headers) > 0 {
		t = t.StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
			}
			return lipgloss.NewStyle()
		})
	}

	// Render the table
	fmt.Fprint(w.config.Writer, t.Render())
	fmt.Fprintln(w.config.Writer) // Add final newline

	return nil
}

