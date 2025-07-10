package tableprinter

import (
	"io"
	"time"

	"github.com/needmore/bc4/internal/tableprinter"
)

// TablePrinter provides bc4-specific table functionality wrapping the core tableprinter
type TablePrinter struct {
	core   tableprinter.TablePrinter
	cs     *tableprinter.ColorScheme
	isTTY  bool
	writer io.Writer
}

// New creates a new bc4 table printer with automatic TTY detection
func New(writer io.Writer) *TablePrinter {
	isTTY := tableprinter.IsTTY(writer)
	maxWidth := tableprinter.GetTerminalWidth()

	return &TablePrinter{
		core:   tableprinter.New(writer, isTTY, maxWidth),
		cs:     tableprinter.NewColorScheme(),
		isTTY:  isTTY,
		writer: writer,
	}
}

// NewWithOptions creates a table printer with specific options
func NewWithOptions(writer io.Writer, isTTY bool, maxWidth int) *TablePrinter {
	return &TablePrinter{
		core:   tableprinter.New(writer, isTTY, maxWidth),
		cs:     tableprinter.NewColorScheme(),
		isTTY:  isTTY,
		writer: writer,
	}
}

// AddHeader adds headers to the table, following GitHub CLI's pattern
func (t *TablePrinter) AddHeader(columns ...string) {
	t.core.AddHeader(columns)
}

// AddField adds a field to the current row
func (t *TablePrinter) AddField(text string, colorFunc ...func(string) string) {
	if len(colorFunc) > 0 {
		t.core.AddField(text, tableprinter.WithColor(colorFunc[0]))
	} else {
		t.core.AddField(text)
	}
}

// AddColorField adds a field with color based on state
func (t *TablePrinter) AddColorField(text, state string) {
	colorFunc := t.cs.ColorFromString(state)
	t.core.AddField(text, tableprinter.WithColor(colorFunc))
}

// AddTimeField adds a time field with GitHub CLI-style formatting
func (t *TablePrinter) AddTimeField(now, timestamp time.Time) {
	var timeStr string

	if t.isTTY {
		// Human-readable relative time for TTY
		timeStr = formatRelativeTime(now, timestamp)
	} else {
		// RFC3339 format for non-TTY (machine readable)
		timeStr = timestamp.Format(time.RFC3339)
	}

	t.core.AddField(timeStr, tableprinter.WithColor(t.cs.Muted))
}

// AddIDField adds an ID field with appropriate formatting
func (t *TablePrinter) AddIDField(id string, state string) {
	// Add # prefix for TTY mode like GitHub CLI
	var displayID string
	if t.isTTY {
		displayID = "#" + id
	} else {
		displayID = id
	}

	colorFunc := t.cs.ColorFromString(state)
	t.core.AddField(displayID, tableprinter.WithColor(colorFunc))
}

// AddProjectField adds a project name field
func (t *TablePrinter) AddProjectField(name, state string) {
	var colorFunc func(string) string
	switch state {
	case "active":
		colorFunc = t.cs.ProjectActive
	case "archived":
		colorFunc = t.cs.ProjectArchived
	default:
		colorFunc = t.cs.Cyan
	}

	t.core.AddField(name, tableprinter.WithColor(colorFunc))
}

// AddTodoField adds a todo item field with status indication
func (t *TablePrinter) AddTodoField(title string, completed bool) {
	var colorFunc func(string) string
	if completed {
		colorFunc = t.cs.TodoCompleted
	} else {
		colorFunc = t.cs.TodoIncomplete
	}

	t.core.AddField(title, tableprinter.WithColor(colorFunc))
}

// AddStatusField adds a status symbol field (like GitHub CLI's check marks)
func (t *TablePrinter) AddStatusField(completed bool) {
	var symbol string
	var colorFunc func(string) string

	if completed {
		symbol = "✓"
		colorFunc = t.cs.Green
	} else {
		symbol = "○"
		colorFunc = t.cs.Gray
	}

	t.core.AddField(symbol, tableprinter.WithColor(colorFunc))
}

// EndRow completes the current row
func (t *TablePrinter) EndRow() {
	t.core.EndRow()
}

// Render outputs the complete table
func (t *TablePrinter) Render() error {
	return t.core.Render()
}

// GetColorScheme returns the color scheme for direct access
func (t *TablePrinter) GetColorScheme() *tableprinter.ColorScheme {
	return t.cs
}

// IsTTY returns whether the output is a terminal
func (t *TablePrinter) IsTTY() bool {
	return t.isTTY
}
