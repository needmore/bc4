package utils

import (
	"strings"
	"testing"
	"time"

	"github.com/needmore/bc4/internal/api"
)

func TestFormatCommentsForDisplay(t *testing.T) {
	t.Run("Empty comments", func(t *testing.T) {
		result, err := FormatCommentsForDisplay([]api.Comment{})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result != "" {
			t.Errorf("Expected empty string, got %q", result)
		}
	})

	t.Run("Single comment", func(t *testing.T) {
		comments := []api.Comment{
			{
				ID:        123,
				Content:   "This is a test comment",
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				Creator: api.Person{
					Name: "John Doe",
				},
			},
		}

		result, err := FormatCommentsForDisplay(comments)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Check that the result contains expected elements
		if !strings.Contains(result, "Comments (1)") {
			t.Errorf("Expected title with count, got %q", result)
		}
		if !strings.Contains(result, "Comment #123") {
			t.Errorf("Expected comment ID, got %q", result)
		}
		if !strings.Contains(result, "John Doe") {
			t.Errorf("Expected author name, got %q", result)
		}
		if !strings.Contains(result, "This is a test comment") {
			t.Errorf("Expected comment content, got %q", result)
		}
	})

	t.Run("Multiple comments", func(t *testing.T) {
		comments := []api.Comment{
			{
				ID:        1,
				Content:   "First comment",
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				Creator: api.Person{
					Name: "Alice",
				},
			},
			{
				ID:        2,
				Content:   "Second comment",
				CreatedAt: time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC),
				Creator: api.Person{
					Name: "Bob",
				},
			},
		}

		result, err := FormatCommentsForDisplay(comments)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Check that both comments are present
		if !strings.Contains(result, "Comments (2)") {
			t.Errorf("Expected count of 2, got %q", result)
		}
		if !strings.Contains(result, "First comment") {
			t.Errorf("Expected first comment, got %q", result)
		}
		if !strings.Contains(result, "Second comment") {
			t.Errorf("Expected second comment, got %q", result)
		}
		if !strings.Contains(result, "Alice") {
			t.Errorf("Expected first author, got %q", result)
		}
		if !strings.Contains(result, "Bob") {
			t.Errorf("Expected second author, got %q", result)
		}
	})
}
