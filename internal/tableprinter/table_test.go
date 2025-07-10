package tableprinter

import (
	"bytes"
	"strings"
	"testing"
)

func TestTTYTablePrinter(t *testing.T) {
	var buf bytes.Buffer

	// Create TTY table printer
	printer := New(&buf, true, 80)

	// Add headers
	printer.AddHeader([]string{"ID", "NAME", "STATUS"})

	// Add rows
	printer.AddField("123")
	printer.AddField("Test Project")
	printer.AddField("Active")
	printer.EndRow()

	printer.AddField("456")
	printer.AddField("Another Project")
	printer.AddField("Archived")
	printer.EndRow()

	// Render
	err := printer.Render()
	if err != nil {
		t.Fatalf("Failed to render: %v", err)
	}

	output := buf.String()

	// Basic validation
	if !strings.Contains(output, "Test Project") {
		t.Error("Output should contain 'Test Project'")
	}

	if !strings.Contains(output, "Active") {
		t.Error("Output should contain 'Active'")
	}

	// Should have proper spacing (3 spaces between columns)
	if !strings.Contains(output, "123   Test Project") {
		t.Error("Output should have proper column spacing")
	}
}

func TestTSVTablePrinter(t *testing.T) {
	var buf bytes.Buffer

	// Create TSV table printer (non-TTY)
	printer := New(&buf, false, 80)

	// Add headers
	printer.AddHeader([]string{"ID", "NAME", "STATUS"})

	// Add rows
	printer.AddField("123")
	printer.AddField("Test Project")
	printer.AddField("Active")
	printer.EndRow()

	// Render
	err := printer.Render()
	if err != nil {
		t.Fatalf("Failed to render: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should have header and data row
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(lines))
	}

	// Header should be tab-separated
	if lines[0] != "ID\tNAME\tSTATUS" {
		t.Errorf("Expected tab-separated header, got: %s", lines[0])
	}

	// Data should be tab-separated
	if lines[1] != "123\tTest Project\tActive" {
		t.Errorf("Expected tab-separated data, got: %s", lines[1])
	}
}

func TestColumnWidthCalculation(t *testing.T) {
	var buf bytes.Buffer
	printer := &ttyTablePrinter{
		writer:   &buf,
		maxWidth: 40, // Narrow terminal
	}

	// Set up content for width calculation
	printer.columnContent = [][]string{
		{"ID", "123", "456"},
		{"NAME", "Very Long Project Name That Should Be Truncated", "Short"},
		{"STATUS", "Active", "Archived"},
	}

	printer.calculateColumnWidths()

	// Should have 3 columns
	if len(printer.columnWidths) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(printer.columnWidths))
	}

	// Total width should not exceed available space
	totalWidth := 0
	for _, width := range printer.columnWidths {
		totalWidth += width
	}

	// Add separator space (2 separators * 3 chars each = 6)
	totalWidth += 6

	if totalWidth > 40 {
		t.Errorf("Total width %d exceeds max width 40", totalWidth)
	}

	// All columns should have minimum width
	for i, width := range printer.columnWidths {
		if width < 8 {
			t.Errorf("Column %d width %d is below minimum 8", i, width)
		}
	}
}

func TestColorScheme(t *testing.T) {
	cs := NewColorScheme()

	// Test basic color functions exist
	if cs.Green == nil {
		t.Error("Green color function should not be nil")
	}

	if cs.Red == nil {
		t.Error("Red color function should not be nil")
	}

	// Test ColorFromString
	greenFunc := cs.ColorFromString("active")
	if greenFunc == nil {
		t.Error("Should return a color function for 'active' state")
	}

	redFunc := cs.ColorFromString("completed")
	if redFunc == nil {
		t.Error("Should return a color function for 'completed' state")
	}
}
