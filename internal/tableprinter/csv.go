package tableprinter

import (
	"encoding/csv"
	"io"
)

// csvTablePrinter implements TablePrinter for comma-separated output (scripts/pipes)
type csvTablePrinter struct {
	writer    io.Writer
	csvWriter *csv.Writer
	headers   []string
	rows      [][]string

	// Current row being built
	currentRow []string
}

func (c *csvTablePrinter) AddHeader(columns []string, opts ...fieldOption) {
	// For CSV output, we ignore field options and just store the text
	c.headers = make([]string, len(columns))
	copy(c.headers, columns)
}

func (c *csvTablePrinter) AddField(text string, opts ...fieldOption) {
	// For CSV output, strip ANSI codes and ignore formatting options
	cleanText := stripAnsi(text)
	c.currentRow = append(c.currentRow, cleanText)
}

func (c *csvTablePrinter) EndRow() {
	if len(c.currentRow) > 0 {
		c.rows = append(c.rows, c.currentRow)
		c.currentRow = nil
	}
}

func (c *csvTablePrinter) Render() error {
	// Initialize CSV writer if not already done
	if c.csvWriter == nil {
		c.csvWriter = csv.NewWriter(c.writer)
	}

	// Write headers if present
	if len(c.headers) > 0 {
		// Check if we have any non-empty headers
		hasContent := false
		for _, h := range c.headers {
			if h != "" {
				hasContent = true
				break
			}
		}

		if hasContent {
			if err := c.csvWriter.Write(c.headers); err != nil {
				return err
			}
		}
	}

	// Write data rows
	for _, row := range c.rows {
		if err := c.csvWriter.Write(row); err != nil {
			return err
		}
	}

	// Flush to ensure all data is written
	c.csvWriter.Flush()
	return c.csvWriter.Error()
}
