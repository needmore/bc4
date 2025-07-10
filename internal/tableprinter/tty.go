package tableprinter

import (
	"fmt"
	"io"
	"strings"
)

// ttyTablePrinter implements TablePrinter for terminal output with formatting
type ttyTablePrinter struct {
	writer   io.Writer
	maxWidth int
	headers  []field
	rows     [][]field

	// Current row being built
	currentRow []field

	// Calculated column widths
	columnWidths []int

	// Content tracking for width calculation
	columnContent [][]string // [column][content] for measuring
}

func (t *ttyTablePrinter) AddHeader(columns []string, opts ...fieldOption) {
	t.headers = make([]field, len(columns))
	for i, col := range columns {
		f := &field{Text: col}
		for _, opt := range opts {
			opt(f)
		}
		t.headers[i] = *f
	}

	// Initialize column content tracking
	t.columnContent = make([][]string, len(columns))
	for i, col := range columns {
		t.columnContent[i] = []string{col} // Start with header content
	}
}

func (t *ttyTablePrinter) AddField(text string, opts ...fieldOption) {
	f := &field{Text: text}
	for _, opt := range opts {
		opt(f)
	}

	t.currentRow = append(t.currentRow, *f)

	// Track content for width calculation
	colIndex := len(t.currentRow) - 1
	if colIndex < len(t.columnContent) {
		t.columnContent[colIndex] = append(t.columnContent[colIndex], text)
	}
}

func (t *ttyTablePrinter) EndRow() {
	if len(t.currentRow) > 0 {
		t.rows = append(t.rows, t.currentRow)
		t.currentRow = nil
	}
}

func (t *ttyTablePrinter) Render() error {
	if len(t.rows) == 0 && len(t.headers) == 0 {
		return nil
	}

	// Calculate optimal column widths using GitHub CLI's algorithm
	t.calculateColumnWidths()

	// Render headers if present
	if len(t.headers) > 0 {
		t.renderRow(t.headers, true)
	}

	// Render data rows
	for _, row := range t.rows {
		t.renderRow(row, false)
	}

	return nil
}

// calculateColumnWidths implements GitHub CLI's intelligent width distribution algorithm
func (t *ttyTablePrinter) calculateColumnWidths() {
	numCols := len(t.columnContent)
	if numCols == 0 {
		return
	}

	// Step 1: Measure natural width of each column
	naturalWidths := make([]int, numCols)
	for i, colContent := range t.columnContent {
		maxWidth := 0
		for _, content := range colContent {
			width := measureWidth(content)
			if width > maxWidth {
				maxWidth = width
			}
		}
		naturalWidths[i] = maxWidth
	}

	// Step 2: Calculate separator overhead (GitHub CLI uses 3 chars between columns)
	separatorWidth := (numCols - 1) * 3
	availableWidth := t.maxWidth - separatorWidth

	// Step 3: Check if natural widths fit
	totalNaturalWidth := 0
	for _, width := range naturalWidths {
		totalNaturalWidth += width
	}

	if totalNaturalWidth <= availableWidth {
		// Everything fits naturally - distribute any extra space to the last column
		t.columnWidths = make([]int, numCols)
		copy(t.columnWidths, naturalWidths)

		extraSpace := availableWidth - totalNaturalWidth
		if extraSpace > 0 && numCols > 0 {
			t.columnWidths[numCols-1] += extraSpace
		}
		return
	}

	// Step 4: Need to truncate - use GitHub CLI's proportional algorithm
	t.columnWidths = make([]int, numCols)

	// Set minimum widths (GitHub CLI uses 8 as minimum)
	minWidth := 8
	totalMinWidth := numCols * minWidth

	if totalMinWidth > availableWidth {
		// Terminal too narrow - just use minimum widths
		for i := range t.columnWidths {
			t.columnWidths[i] = minWidth
		}
		return
	}

	// Proportionally distribute available width above minimums
	extraSpace := availableWidth - totalMinWidth

	// Calculate weights based on natural widths
	totalWeight := 0
	weights := make([]int, numCols)
	for i, width := range naturalWidths {
		weight := width - minWidth
		if weight < 0 {
			weight = 0
		}
		weights[i] = weight
		totalWeight += weight
	}

	// Distribute extra space proportionally
	for i := range t.columnWidths {
		t.columnWidths[i] = minWidth
		if totalWeight > 0 {
			extraForCol := (extraSpace * weights[i]) / totalWeight
			t.columnWidths[i] += extraForCol
		}
	}

	// Handle any remaining space due to integer division
	remaining := availableWidth
	for _, width := range t.columnWidths {
		remaining -= width
	}

	// Give any remaining space to the last column
	if remaining > 0 && numCols > 0 {
		t.columnWidths[numCols-1] += remaining
	}
}

// renderRow renders a single row with proper formatting
func (t *ttyTablePrinter) renderRow(row []field, isHeader bool) {
	if len(row) == 0 {
		return
	}

	var parts []string

	for i, f := range row {
		var content string
		var width int

		if i < len(t.columnWidths) {
			width = t.columnWidths[i]
		} else {
			width = 20 // fallback width
		}

		// Apply truncation if needed
		if f.TruncateFunc != nil {
			content = f.TruncateFunc(width, f.Text)
		} else {
			content = defaultTruncate(width, f.Text)
		}

		// Apply color if specified
		if f.ColorFunc != nil {
			content = f.ColorFunc(content)
		}

		// Apply padding
		if f.PaddingFunc != nil {
			content = f.PaddingFunc(width, content)
		} else {
			content = defaultPadding(width, content)
		}

		parts = append(parts, content)
	}

	// Join with GitHub CLI-style spacing (3 spaces)
	output := strings.Join(parts, "   ")
	fmt.Fprintln(t.writer, output)
}
