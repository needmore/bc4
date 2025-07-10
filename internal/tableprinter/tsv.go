package tableprinter

import (
	"fmt"
	"io"
	"strings"
)

// tsvTablePrinter implements TablePrinter for tab-separated output (scripts/pipes)
type tsvTablePrinter struct {
	writer  io.Writer
	headers []string
	rows    [][]string

	// Current row being built
	currentRow []string
}

func (t *tsvTablePrinter) AddHeader(columns []string, opts ...fieldOption) {
	// For TSV output, we ignore field options and just store the text
	t.headers = make([]string, len(columns))
	copy(t.headers, columns)
}

func (t *tsvTablePrinter) AddField(text string, opts ...fieldOption) {
	// For TSV output, strip ANSI codes and ignore formatting options
	cleanText := stripAnsi(text)
	t.currentRow = append(t.currentRow, cleanText)
}

func (t *tsvTablePrinter) EndRow() {
	if len(t.currentRow) > 0 {
		t.rows = append(t.rows, t.currentRow)
		t.currentRow = nil
	}
}

func (t *tsvTablePrinter) Render() error {
	// Render headers if present
	if len(t.headers) > 0 {
		// Check if we have any non-empty headers
		hasContent := false
		for _, h := range t.headers {
			if h != "" {
				hasContent = true
				break
			}
		}

		if hasContent {
			fmt.Fprintln(t.writer, strings.Join(t.headers, "\t"))
		}
	}

	// Render data rows
	for _, row := range t.rows {
		fmt.Fprintln(t.writer, strings.Join(row, "\t"))
	}

	return nil
}
