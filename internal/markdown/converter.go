package markdown

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// Converter handles conversion between Markdown and Basecamp's rich text format
type Converter interface {
	MarkdownToRichText(markdown string) (string, error)
	RichTextToMarkdown(richtext string) (string, error)
}

// converter implements the Converter interface
type converter struct {
	md goldmark.Markdown
}

// NewConverter creates a new markdown converter
func NewConverter() Converter {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithXHTML(),
		),
	)

	return &converter{md: md}
}

// MarkdownToRichText converts GitHub Flavored Markdown to Basecamp's rich text HTML format
func (c *converter) MarkdownToRichText(markdown string) (string, error) {
	var buf bytes.Buffer
	if err := c.md.Convert([]byte(markdown), &buf); err != nil {
		return "", fmt.Errorf("failed to convert markdown: %w", err)
	}

	// Get the HTML output
	html := buf.String()

	// Post-process the HTML to match Basecamp's format
	html = c.postProcessHTML(html)

	// Clean up the output
	result := strings.TrimSpace(html)

	// Handle empty input
	if result == "" || result == "<div></div>" {
		return "", nil
	}

	return result, nil
}

// postProcessHTML transforms standard HTML to Basecamp's rich text format
func (c *converter) postProcessHTML(html string) string {
	// Replace <p> tags with <div> tags
	html = strings.ReplaceAll(html, "<p>", "<div>")
	html = strings.ReplaceAll(html, "</p>", "</div>")

	// Replace all heading levels with h1
	for i := 2; i <= 6; i++ {
		html = strings.ReplaceAll(html, fmt.Sprintf("<h%d>", i), "<h1>")
		html = strings.ReplaceAll(html, fmt.Sprintf("</h%d>", i), "</h1>")
		// Also handle headings with id attributes
		re := regexp.MustCompile(fmt.Sprintf(`<h%d[^>]*>`, i))
		html = re.ReplaceAllString(html, "<h1>")
	}

	// Replace <del> with <strike> for strikethrough
	html = strings.ReplaceAll(html, "<del>", "<strike>")
	html = strings.ReplaceAll(html, "</del>", "</strike>")

	// Replace <code> with <pre> for inline code
	html = strings.ReplaceAll(html, "<code>", "<pre>")
	html = strings.ReplaceAll(html, "</code>", "</pre>")

	// Clean up code blocks - fix double wrapping
	re := regexp.MustCompile(`<pre><code[^>]*>`)
	html = re.ReplaceAllString(html, "<pre>")
	html = strings.ReplaceAll(html, "</code></pre>", "</pre>")
	// Fix double pre tags from inline code conversion
	html = strings.ReplaceAll(html, "<pre><pre>", "<pre>")
	html = strings.ReplaceAll(html, "</pre></pre>", "</pre>")

	// Remove <hr /> (horizontal rules) and replace with <br>
	html = strings.ReplaceAll(html, "<hr />", "<br>\n")
	html = strings.ReplaceAll(html, "<hr/>", "<br>\n")
	html = strings.ReplaceAll(html, "<hr>", "<br>\n")

	// Convert XHTML style breaks to HTML style
	html = strings.ReplaceAll(html, "<br />", "<br>")

	// Clean up list formatting - ensure newlines are consistent
	html = c.cleanListFormatting(html)

	// Remove any remaining unsupported tags (like span, etc.)
	html = c.stripUnsupportedTags(html)

	// Fix quotes in attributes (&#39; -> ')
	html = strings.ReplaceAll(html, "&#39;", "'")

	// Remove HTML comments
	html = regexp.MustCompile(`<!-- [^>]* -->`).ReplaceAllString(html, "")

	// Clean up blockquote formatting
	html = strings.ReplaceAll(html, "<blockquote>\n", "<blockquote>")
	html = strings.ReplaceAll(html, "\n</blockquote>", "</blockquote>")

	// Clean up excessive newlines
	re = regexp.MustCompile(`\n{3,}`)
	html = re.ReplaceAllString(html, "\n\n")

	// Final cleanup - remove newlines within line breaks
	html = strings.ReplaceAll(html, "<br>\n", "<br>")

	return html
}

// cleanListFormatting ensures lists have proper newlines
func (c *converter) cleanListFormatting(html string) string {
	// Fix nested lists first
	html = strings.ReplaceAll(html, "</li><ul>", "</li>\n<ul>")
	html = strings.ReplaceAll(html, "</li><ol>", "</li>\n<ol>")
	html = strings.ReplaceAll(html, "</ul></li>", "</ul>\n</li>")
	html = strings.ReplaceAll(html, "</ol></li>", "</ol>\n</li>")

	// Add newlines after list closures if not at end of string
	html = strings.ReplaceAll(html, "</ul><", "</ul>\n<")
	html = strings.ReplaceAll(html, "</ol><", "</ol>\n<")

	// Add newlines after list items if followed by another list item
	html = strings.ReplaceAll(html, "</li><li>", "</li>\n<li>")

	// Remove trailing newline after final list
	html = strings.TrimRight(html, "\n")
	if strings.HasSuffix(html, "</ul>") || strings.HasSuffix(html, "</ol>") {
		// Add it back if needed
		if !strings.HasSuffix(html, "</ul>") && !strings.HasSuffix(html, "</ol>") {
			html += "\n"
		}
	}

	return html
}

// stripUnsupportedTags removes HTML tags not supported by Basecamp
func (c *converter) stripUnsupportedTags(html string) string {
	// For now, just remove specific known unsupported tags
	// Remove span tags
	html = regexp.MustCompile(`<span[^>]*>`).ReplaceAllString(html, "")
	html = strings.ReplaceAll(html, "</span>", "")

	// Remove any remaining style attributes
	html = regexp.MustCompile(` style="[^"]*"`).ReplaceAllString(html, "")
	html = regexp.MustCompile(` class="[^"]*"`).ReplaceAllString(html, "")

	// Remove id attributes from headings (goldmark adds them)
	html = regexp.MustCompile(` id="[^"]*"`).ReplaceAllString(html, "")

	// Clean up "raw HTML omitted" messages
	html = strings.ReplaceAll(html, "<!-- raw HTML omitted -->", "")
	html = strings.ReplaceAll(html, "raw HTML", "")

	return html
}

// RichTextToMarkdown converts Basecamp's rich text to Markdown
func (c *converter) RichTextToMarkdown(richtext string) (string, error) {
	// Handle empty input
	if richtext == "" || richtext == "<div></div>" {
		return "", nil
	}

	// For now, provide a simple implementation that handles basic cases
	// A proper implementation would use an HTML parser

	// Replace div with p for consistency
	html := strings.ReplaceAll(richtext, "<div>", "<p>")
	html = strings.ReplaceAll(html, "</div>", "</p>")

	// Remove all HTML tags for a basic conversion
	// This is a simplified implementation
	result := html

	// Handle specific tags
	result = regexp.MustCompile(`<h1[^>]*>`).ReplaceAllString(result, "# ")
	result = strings.ReplaceAll(result, "</h1>", "\n\n")
	result = strings.ReplaceAll(result, "<p>", "")
	result = strings.ReplaceAll(result, "</p>", "\n\n")
	result = strings.ReplaceAll(result, "<strong>", "**")
	result = strings.ReplaceAll(result, "</strong>", "**")
	result = strings.ReplaceAll(result, "<b>", "**")
	result = strings.ReplaceAll(result, "</b>", "**")
	result = strings.ReplaceAll(result, "<em>", "*")
	result = strings.ReplaceAll(result, "</em>", "*")
	result = strings.ReplaceAll(result, "<i>", "*")
	result = strings.ReplaceAll(result, "</i>", "*")
	result = strings.ReplaceAll(result, "<strike>", "~~")
	result = strings.ReplaceAll(result, "</strike>", "~~")
	result = strings.ReplaceAll(result, "<del>", "~~")
	result = strings.ReplaceAll(result, "</del>", "~~")
	result = strings.ReplaceAll(result, "<br>", "\n")
	result = strings.ReplaceAll(result, "<br/>", "\n")
	result = strings.ReplaceAll(result, "<br />", "\n")

	// Handle lists
	result = strings.ReplaceAll(result, "<ul>", "")
	result = strings.ReplaceAll(result, "</ul>", "\n")
	result = strings.ReplaceAll(result, "<li>", "- ")
	result = strings.ReplaceAll(result, "</li>", "\n")

	// Handle blockquotes
	result = strings.ReplaceAll(result, "<blockquote>", "> ")
	result = strings.ReplaceAll(result, "</blockquote>", "\n\n")

	// Handle pre tags - check context to determine if inline or block
	// Look for pre tags that are clearly inline (surrounded by other content on same line)
	if regexp.MustCompile(`[^>\s]\s*<pre>`).MatchString(result) || regexp.MustCompile(`</pre>\s*[^<\s]`).MatchString(result) {
		// Inline code
		result = strings.ReplaceAll(result, "<pre>", "`")
		result = strings.ReplaceAll(result, "</pre>", "`")
	} else {
		// Code block
		result = regexp.MustCompile(`<pre>\s*`).ReplaceAllString(result, "```\n")
		result = regexp.MustCompile(`\s*</pre>`).ReplaceAllString(result, "\n```")
	}

	// Decode HTML entities
	result = strings.ReplaceAll(result, "&amp;", "&")
	result = strings.ReplaceAll(result, "&lt;", "<")
	result = strings.ReplaceAll(result, "&gt;", ">")
	result = strings.ReplaceAll(result, "&quot;", "\"")
	result = strings.ReplaceAll(result, "&#39;", "'")
	result = strings.ReplaceAll(result, "&nbsp;", " ")

	// Clean up multiple newlines
	result = regexp.MustCompile(`\n{3,}`).ReplaceAllString(result, "\n\n")

	// Trim the result
	result = strings.TrimSpace(result)

	return result, nil
}
