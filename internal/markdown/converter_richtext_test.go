package markdown

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRichTextToMarkdown(t *testing.T) {
	converter := NewConverter()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "empty div",
			input:    "<div></div>",
			expected: "",
		},
		{
			name:     "simple paragraph",
			input:    "<div>Hello World</div>",
			expected: "Hello World",
		},
		{
			name:     "multiple paragraphs",
			input:    "<div>First paragraph</div><div>Second paragraph</div>",
			expected: "First paragraph\n\nSecond paragraph",
		},
		{
			name:     "heading",
			input:    "<h1>Main Title</h1><div>Content below</div>",
			expected: "# Main Title\n\nContent below",
		},
		{
			name:     "bold text",
			input:    "<div>This is <strong>bold</strong> text</div>",
			expected: "This is **bold** text",
		},
		{
			name:     "italic text",
			input:    "<div>This is <em>italic</em> text</div>",
			expected: "This is *italic* text",
		},
		{
			name:     "strikethrough text",
			input:    "<div>This is <strike>deleted</strike> text</div>",
			expected: "This is ~~deleted~~ text",
		},
		{
			name:     "inline code",
			input:    "<div>Use <pre>console.log()</pre> to debug</div>",
			expected: "Use `console.log()` to debug",
		},
		{
			name:     "code block",
			input:    "<pre>function hello() {\n  console.log('Hello');\n}</pre>",
			expected: "```\nfunction hello() {\n  console.log('Hello');\n}\n```",
		},
		{
			name:     "unordered list",
			input:    "<ul><li>First item</li><li>Second item</li></ul>",
			expected: "- First item\n- Second item",
		},
		{
			name:     "blockquote",
			input:    "<blockquote>This is a quote</blockquote>",
			expected: "> This is a quote",
		},
		{
			name:     "line break",
			input:    "<div>Line one<br>Line two</div>",
			expected: "Line one\nLine two",
		},
		{
			name:     "mixed formatting",
			input:    "<div>This has <strong>bold</strong> and <em>italic</em> and <pre>code</pre></div>",
			expected: "This has **bold** and *italic* and `code`",
		},
		{
			name:     "HTML entities",
			input:    "<div>Quotes: &quot;Hello&quot; &amp; &lt;tags&gt;</div>",
			expected: `Quotes: "Hello" & <tags>`,
		},
		{
			name:  "complex example",
			input: `<h1>Welcome</h1><div>This is a <strong>test</strong> message.</div><ul><li>Item 1</li><li>Item 2</li></ul><div>End</div>`,
			expected: `# Welcome

This is a **test** message.

- Item 1
- Item 2

End`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.RichTextToMarkdown(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
