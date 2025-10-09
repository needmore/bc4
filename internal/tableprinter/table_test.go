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

func TestCSVTablePrinter(t *testing.T) {
	var buf bytes.Buffer

	// Create CSV table printer (non-TTY)
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

	// Header should be comma-separated
	if lines[0] != "ID,NAME,STATUS" {
		t.Errorf("Expected comma-separated header, got: %s", lines[0])
	}

	// Data should be comma-separated
	if lines[1] != "123,Test Project,Active" {
		t.Errorf("Expected comma-separated data, got: %s", lines[1])
	}
}

func TestCSVTablePrinterWithEscaping(t *testing.T) {
	var buf bytes.Buffer

	// Create CSV table printer (non-TTY)
	printer := New(&buf, false, 80)

	// Add headers
	printer.AddHeader([]string{"ID", "NAME", "DESCRIPTION"})

	// Add rows with special characters that need escaping
	printer.AddField("123")
	printer.AddField("Project, with comma")
	printer.AddField("Description with \"quotes\"")
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

	// Header should be comma-separated
	if lines[0] != "ID,NAME,DESCRIPTION" {
		t.Errorf("Expected comma-separated header, got: %s", lines[0])
	}

	// Data should be properly escaped
	expected := "123,\"Project, with comma\",\"Description with \"\"quotes\"\"\""
	if lines[1] != expected {
		t.Errorf("Expected escaped CSV data, got: %s", lines[1])
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

func TestEmojiHandling(t *testing.T) {
	var buf bytes.Buffer

	// Create TTY table printer with narrow width to force truncation
	printer := New(&buf, true, 60)

	// Add headers
	printer.AddHeader([]string{"ID", "NAME", "DESCRIPTION"})

	// Add row with emojis that caused the panic
	// The emoji "👩‍🎨" is actually composed of multiple Unicode code points
	printer.AddField("1477")
	printer.AddField("A - Designers 👩‍🎨👨‍🎨")
	printer.AddField("Design team workspace")
	printer.EndRow()

	// This should not panic
	err := printer.Render()
	if err != nil {
		t.Fatalf("Failed to render: %v", err)
	}

	output := buf.String()

	// Should contain some part of the text (may be truncated)
	if !strings.Contains(output, "Designers") {
		t.Error("Output should contain 'Designers'")
	}
}

func TestMultiByteCharacterTruncation(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		maxWidth int
		wantLen  int // expected rune count (approximately)
	}{
		{
			name:     "emoji with zero-width joiners",
			input:    "👩‍🎨👨‍🎨 Test",
			maxWidth: 10,
			wantLen:  10,
		},
		{
			name:     "mixed ascii and emoji",
			input:    "Hello 世界 🌍",
			maxWidth: 12,
			wantLen:  12,
		},
		{
			name:     "only emojis",
			input:    "🎨🎭🎪🎬🎮",
			maxWidth: 5,
			wantLen:  5,
		},
		{
			name:     "very narrow width",
			input:    "A - Designers 👩‍🎨👨‍🎨",
			maxWidth: 8,
			wantLen:  8,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := defaultTruncate(tc.maxWidth, tc.input)

			// Should not panic
			if result == "" {
				t.Error("Result should not be empty")
			}

			// Result should not exceed maxWidth in display width
			width := measureWidth(result)
			if width > tc.maxWidth {
				t.Errorf("Result width %d exceeds maxWidth %d", width, tc.maxWidth)
			}
		})
	}
}

func TestIssue_PanicWithEmojiInProjectList(t *testing.T) {
	// Reproduces the exact scenario from the bug report:
	// Project name: "A - Designers 👩‍🎨👨‍🎨"
	// This caused: "panic: runtime error: slice bounds out of range [:35] with capacity 32"

	var buf bytes.Buffer
	printer := New(&buf, true, 120)

	printer.AddHeader([]string{"ID", "NAME", "DESCRIPTION", "UPDATED"})

	// Add the exact project that caused the panic
	printer.AddField("needmore/bc4#1477...")
	printer.AddField("A - Designers 👩‍🎨👨‍🎨")
	printer.AddField("")
	printer.AddField("2025-08-15T07:4...")
	printer.EndRow()

	printer.AddField("needmore/bc4#2211...")
	printer.AddField("A – Account Managers")
	printer.AddField("PAM c'est par ici !")
	printer.AddField("2025-10-02T11:2...")
	printer.EndRow()

	// This should not panic anymore
	err := printer.Render()
	if err != nil {
		t.Fatalf("Failed to render: %v", err)
	}

	output := buf.String()

	// Verify the emoji project name is in the output
	if !strings.Contains(output, "Designers") {
		t.Error("Output should contain project with emoji in name")
	}
}
