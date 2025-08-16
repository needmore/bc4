package markdown

import (
	"fmt"
	"testing"
)

// TestTodoTitleWrapping tests the current behavior with simple todo titles
func TestTodoTitleWrapping(t *testing.T) {
	converter := NewConverter()
	
	// Test simple todo titles that should NOT be wrapped in DIVs according to the issue
	testCases := []string{
		"Review pull request",
		"Deploy to production", 
		"Fix bug in parser",
		"Update documentation",
		"Simple task title",
	}
	
	fmt.Println("Testing simple todo titles:")
	for _, title := range testCases {
		result, err := converter.MarkdownToRichText(title)
		if err != nil {
			t.Errorf("Error with '%s': %v", title, err)
			continue
		}
		fmt.Printf("Input: %q\n", title)
		fmt.Printf("Output: %q\n", result)
		fmt.Printf("Has DIV wrapper: %t\n\n", result != title)
	}
	
	// Test cases that SHOULD have HTML formatting
	complexCases := []string{
		"Review **critical** pull request",
		"Fix bug\n\nWith additional details",
		"Task with `code` formatting",
		"# Heading task",
	}
	
	fmt.Println("Testing complex content that should keep HTML:")
	for _, content := range complexCases {
		result, err := converter.MarkdownToRichText(content)
		if err != nil {
			t.Errorf("Error with '%s': %v", content, err)
			continue
		}
		fmt.Printf("Input: %q\n", content)
		fmt.Printf("Output: %q\n\n", result)
	}
}