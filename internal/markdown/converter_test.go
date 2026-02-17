package markdown

import (
	"strings"
	"testing"
)

func TestMarkdownToRichText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Basic formatting
		{
			name:     "plain text in paragraph",
			input:    "Hello world",
			expected: "Hello world",
		},
		{
			name:     "bold text",
			input:    "This is **bold** text",
			expected: "<div>This is <strong>bold</strong> text</div>",
		},
		{
			name:     "italic text",
			input:    "This is *italic* text",
			expected: "<div>This is <em>italic</em> text</div>",
		},
		{
			name:     "strikethrough text",
			input:    "This is ~~strikethrough~~ text",
			expected: "<div>This is <strike>strikethrough</strike> text</div>",
		},
		{
			name:     "mixed formatting",
			input:    "This has **bold**, *italic*, and ~~strike~~ text",
			expected: "<div>This has <strong>bold</strong>, <em>italic</em>, and <strike>strike</strike> text</div>",
		},

		// Headings
		{
			name:     "h1 heading",
			input:    "# Heading 1",
			expected: "<h1>Heading 1</h1>",
		},
		{
			name:     "h2 heading (converts to h1)",
			input:    "## Heading 2",
			expected: "<h1>Heading 2</h1>",
		},
		{
			name:     "h3 heading (converts to h1)",
			input:    "### Heading 3",
			expected: "<h1>Heading 3</h1>",
		},

		// Links
		{
			name:     "inline link",
			input:    "Check out [GitHub](https://github.com)",
			expected: `<div>Check out <a href="https://github.com">GitHub</a></div>`,
		},
		{
			name:     "auto link",
			input:    "Visit <https://github.com>",
			expected: `<div>Visit <a href="https://github.com">https://github.com</a></div>`,
		},
		{
			name:     "link with special characters",
			input:    "Search [here](https://example.com?q=test&lang=en)",
			expected: `<div>Search <a href="https://example.com?q=test&amp;lang=en">here</a></div>`,
		},

		// Code
		{
			name:     "inline code",
			input:    "Run `npm install` to install",
			expected: "<div>Run <pre>npm install</pre> to install</div>",
		},
		{
			name:     "code block",
			input:    "```\necho hello\n```",
			expected: "<pre>echo hello\n</pre>",
		},
		{
			name:     "fenced code block with language",
			input:    "```bash\necho hello\n```",
			expected: "<pre>echo hello\n</pre>",
		},
		{
			name:     "code with special characters",
			input:    "Use `<div>` tags",
			expected: "<div>Use <pre>&lt;div&gt;</pre> tags</div>",
		},

		// Lists
		{
			name:     "unordered list",
			input:    "- Item 1\n- Item 2\n- Item 3",
			expected: "<ul>\n<li>Item 1</li>\n<li>Item 2</li>\n<li>Item 3</li>\n</ul>",
		},
		{
			name:     "ordered list",
			input:    "1. First\n2. Second\n3. Third",
			expected: "<ol>\n<li>First</li>\n<li>Second</li>\n<li>Third</li>\n</ol>",
		},
		{
			name:     "nested list",
			input:    "- Item 1\n  - Nested 1\n  - Nested 2\n- Item 2",
			expected: "<ul>\n<li>Item 1\n<ul>\n<li>Nested 1</li>\n<li>Nested 2</li>\n</ul>\n</li>\n<li>Item 2</li>\n</ul>",
		},
		{
			name:     "list with formatting",
			input:    "- **Bold** item\n- *Italic* item\n- ~~Strike~~ item",
			expected: "<ul>\n<li><strong>Bold</strong> item</li>\n<li><em>Italic</em> item</li>\n<li><strike>Strike</strike> item</li>\n</ul>",
		},

		// Blockquotes
		{
			name:     "simple blockquote",
			input:    "> This is a quote",
			expected: "<blockquote><div>This is a quote</div></blockquote>",
		},
		{
			name:     "multiline blockquote",
			input:    "> Line 1\n> Line 2",
			expected: "<blockquote><div>Line 1<br>Line 2</div></blockquote>",
		},
		{
			name:     "blockquote with formatting",
			input:    "> This has **bold** text",
			expected: "<blockquote><div>This has <strong>bold</strong> text</div></blockquote>",
		},

		// Multiple paragraphs
		{
			name:     "multiple paragraphs",
			input:    "Paragraph 1\n\nParagraph 2",
			expected: "<div>Paragraph 1</div>\n<div>Paragraph 2</div>",
		},
		{
			name:     "heading with paragraph",
			input:    "# Title\n\nThis is content",
			expected: "<h1>Title</h1>\n<div>This is content</div>",
		},

		// Line breaks
		{
			name:     "hard line break",
			input:    "Line 1  \nLine 2",
			expected: "<div>Line 1<br>Line 2</div>",
		},
		{
			name:     "thematic break",
			input:    "Above\n\n---\n\nBelow",
			expected: "<div>Above</div>\n<br>\n<div>Below</div>",
		},

		// Edge cases
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   \n\n   ",
			expected: "",
		},
		{
			name:     "special HTML characters",
			input:    "Less than < and greater than > and & ampersand",
			expected: "<div>Less than &lt; and greater than &gt; and &amp; ampersand</div>",
		},
		{
			name:     "raw HTML (should be ignored)",
			input:    "Text with <span>raw HTML</span> inside",
			expected: "<div>Text with  inside</div>",
		},
		{
			name:     "very long line",
			input:    strings.Repeat("word ", 100),
			expected: strings.TrimSpace(strings.Repeat("word ", 100)),
		},

		// Complex examples
		{
			name: "complex document",
			input: `# Project Title

This is a **description** with *various* formatting.

## Features

- Fast performance
- Easy to use
- Well documented

Here's some ` + "`code`" + ` and a [link](https://example.com).

> Important note: This is ~~deprecated~~ updated.

1. First step
2. Second step
3. Third step`,
			expected: `<h1>Project Title</h1>
<div>This is a <strong>description</strong> with <em>various</em> formatting.</div>
<h1>Features</h1>
<ul>
<li>Fast performance</li>
<li>Easy to use</li>
<li>Well documented</li>
</ul>
<div>Here's some <pre>code</pre> and a <a href="https://example.com">link</a>.</div>
<blockquote><div>Important note: This is <strike>deprecated</strike> updated.</div></blockquote>
<ol>
<li>First step</li>
<li>Second step</li>
<li>Third step</li>
</ol>`,
		},
	}

	converter := NewConverter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.MarkdownToRichText(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("MarkdownToRichText() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestMarkdownToRichText_Errors(t *testing.T) {
	converter := NewConverter()

	// Test with nil input (Go strings can't be nil, but we can test empty)
	result, err := converter.MarkdownToRichText("")
	if err != nil {
		t.Errorf("expected no error for empty input, got: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty result for empty input, got: %q", result)
	}
}

// Benchmark tests
func BenchmarkMarkdownToRichText_Simple(b *testing.B) {
	converter := NewConverter()
	input := "This is **bold** and *italic* text with a [link](https://example.com)."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = converter.MarkdownToRichText(input)
	}
}

func BenchmarkMarkdownToRichText_Complex(b *testing.B) {
	converter := NewConverter()
	input := `# Title

This is a paragraph with **bold**, *italic*, and ~~strikethrough~~ text.

## List

- Item 1
- Item 2
  - Nested item
- Item 3

> Blockquote with ` + "`code`" + ` inside.

1. First
2. Second
3. Third`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = converter.MarkdownToRichText(input)
	}
}
