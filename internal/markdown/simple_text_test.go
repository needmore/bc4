package markdown

import (
	"testing"
)

// TestSimplePlainTextDetection tests the isSimplePlainText detection logic
func TestSimplePlainTextDetection(t *testing.T) {
	converter := &converter{}

	tests := []struct {
		input    string
		expected bool
		reason   string
	}{
		// Should be simple plain text
		{"Simple task", true, "basic text"},
		{"Review pull request", true, "task description"},
		{"Deploy to production", true, "simple action"},
		{"Fix bugs", true, "short task"},
		{"Update the documentation files", true, "longer but simple text"},
		{"A very long task description that goes on and on but contains no markdown formatting at all", true, "very long plain text"},

		// Should NOT be simple plain text
		{"Task with **bold** text", false, "contains bold formatting"},
		{"Task with *italic* text", false, "contains italic formatting"},
		{"Task with `code` snippet", false, "contains code formatting"},
		{"Task with ~~strikethrough~~ text", false, "contains strikethrough"},
		{"Check out [this link](https://example.com)", false, "contains link"},
		{"Visit <https://example.com>", false, "contains URL"},
		{"# Heading task", false, "contains heading"},
		{"Simple task\nWith second line", false, "multiline"},
		{"> Quote something", false, "contains blockquote"},
		{"- Item 1", false, "contains list marker"},
		{"+ Item 1", false, "contains list marker"},
		{"1. First item", false, "contains numbered list"},
		{"Task with <html> tags", false, "contains HTML"},
		{"Task with & entity", false, "contains HTML entity"},
		{"Email me at user@example.com", false, "contains @ symbol"},
		{"", false, "empty string"},
		{"   ", false, "whitespace only"},
	}

	for _, tt := range tests {
		t.Run(tt.reason, func(t *testing.T) {
			result := converter.isSimplePlainText(tt.input)
			if result != tt.expected {
				t.Errorf("isSimplePlainText(%q) = %v, want %v (%s)", tt.input, result, tt.expected, tt.reason)
			}
		})
	}
}
