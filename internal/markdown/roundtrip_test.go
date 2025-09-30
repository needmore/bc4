package markdown

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRoundTripConversion tests that Markdown -> Rich Text -> Markdown preserves content
func TestRoundTripConversion(t *testing.T) {
	converter := NewConverter()

	tests := []struct {
		name  string
		input string
		// Sometimes the round-trip might produce slightly different but equivalent markdown
		expectedOutput string
	}{
		{
			name:           "simple formatting",
			input:          "This has **bold** and *italic* text",
			expectedOutput: "This has **bold** and *italic* text",
		},
		{
			name:           "strikethrough",
			input:          "This is ~~deleted~~ text",
			expectedOutput: "This is ~~deleted~~ text",
		},
		{
			name:           "heading with content",
			input:          "# Main Title\n\nSome content below",
			expectedOutput: "# Main Title\n\nSome content below",
		},
		{
			name:           "unordered list",
			input:          "- First item\n- Second item",
			expectedOutput: "- First item\n- Second item",
		},
		{
			name:           "inline code",
			input:          "Use `console.log()` to debug",
			expectedOutput: "Use `console.log()` to debug",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to rich text
			richText, err := converter.MarkdownToRichText(tt.input)
			assert.NoError(t, err)
			assert.NotEmpty(t, richText)

			// Convert back to markdown
			result, err := converter.RichTextToMarkdown(richText)
			assert.NoError(t, err)

			// For round-trip tests, we allow some formatting differences
			// as long as the semantic content is preserved
			normalizedResult := normalizeMarkdown(result)
			normalizedExpected := normalizeMarkdown(tt.expectedOutput)
			assert.Equal(t, normalizedExpected, normalizedResult)
		})
	}
}

// normalizeMarkdown removes extra whitespace for comparison
func normalizeMarkdown(md string) string {
	// Remove extra newlines between list items
	result := regexp.MustCompile(`\n+- `).ReplaceAllString(md, "\n- ")
	// Normalize multiple newlines
	result = regexp.MustCompile(`\n{3,}`).ReplaceAllString(result, "\n\n")
	return strings.TrimSpace(result)
}

// TestBasecampHTMLValidation tests the HTML validation functionality
func TestBasecampHTMLValidation(t *testing.T) {
	converter := NewConverter()

	tests := []struct {
		name        string
		html        string
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "valid basecamp HTML",
			html:        "<div>Test <strong>bold</strong> text</div>",
			shouldError: false,
		},
		{
			name:        "valid with links",
			html:        `<div>Visit <a href="https://example.com">here</a></div>`,
			shouldError: false,
		},
		{
			name:        "valid with lists",
			html:        "<ul><li>Item 1</li><li>Item 2</li></ul>",
			shouldError: false,
		},
		{
			name:        "invalid tag",
			html:        "<div>Text with <span>invalid</span> tag</div>",
			shouldError: true,
			errorMsg:    "unsupported HTML tag: span",
		},
		{
			name:        "empty input",
			html:        "",
			shouldError: false,
		},
		{
			name:        "bc-attachment tag",
			html:        `<bc-attachment sgid="test">@mention</bc-attachment>`,
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := converter.ValidateBasecampHTML(tt.html)

			if tt.shouldError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestComplexNesting tests complex nested structures
func TestComplexNesting(t *testing.T) {
	converter := NewConverter()

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "deeply nested lists",
			input: "- Item 1\n  - Nested 1\n    - Deep nested\n  - Nested 2\n- Item 2",
		},
		{
			name:  "mixed formatting in lists",
			input: "- **Bold** item\n- *Italic* item with `code`\n- ~~Strike~~ item",
		},
		{
			name:  "blockquote with formatting",
			input: "> This is a quote with **bold** and *italic* text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to rich text
			richText, err := converter.MarkdownToRichText(tt.input)
			assert.NoError(t, err)
			assert.NotEmpty(t, richText)

			// Validate the HTML
			err = converter.ValidateBasecampHTML(richText)
			assert.NoError(t, err)

			// Convert back to markdown (should not error)
			result, err := converter.RichTextToMarkdown(richText)
			assert.NoError(t, err)
			assert.NotEmpty(t, result)
		})
	}
}
