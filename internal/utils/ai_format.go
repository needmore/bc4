package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/markdown"
)

// parseTime attempts to parse a timestamp string in common Basecamp formats
func parseTime(timeStr string) (time.Time, error) {
	// Try RFC3339 format first (most common for Basecamp API)
	if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
		return t, nil
	}
	// Fallback to other common formats
	if t, err := time.Parse("2006-01-02T15:04:05Z", timeStr); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("unable to parse time: %s", timeStr)
}

// FormatCardAsMarkdown formats a card with all its comments as AI-optimized markdown
func FormatCardAsMarkdown(card *api.Card, comments []api.Comment) (string, error) {
	var buf strings.Builder
	converter := markdown.NewConverter()

	// Title
	fmt.Fprintf(&buf, "# %s\n\n", card.Title)

	// Metadata
	fmt.Fprintf(&buf, "**ID:** %d\n", card.ID)
	fmt.Fprintf(&buf, "**Status:** %s\n", card.Status)

	if card.Parent != nil {
		fmt.Fprintf(&buf, "**Column:** %s", card.Parent.Title)
		if card.Parent.Color != "" && card.Parent.Color != "white" {
			fmt.Fprintf(&buf, " (%s)", card.Parent.Color)
		}
		fmt.Fprint(&buf, "\n")
	}

	fmt.Fprintf(&buf, "**Created:** %s\n", card.CreatedAt.Format("2006-01-02 15:04"))
	fmt.Fprintf(&buf, "**Updated:** %s\n", card.UpdatedAt.Format("2006-01-02 15:04"))

	if len(card.Assignees) > 0 {
		assigneeNames := make([]string, len(card.Assignees))
		for i, assignee := range card.Assignees {
			assigneeNames[i] = assignee.Name
		}
		fmt.Fprintf(&buf, "**Assignees:** %s\n", strings.Join(assigneeNames, ", "))
	}

	if card.DueOn != nil && *card.DueOn != "" {
		fmt.Fprintf(&buf, "**Due:** %s\n", *card.DueOn)
	}

	if card.Creator != nil {
		fmt.Fprintf(&buf, "**Created by:** %s\n", card.Creator.Name)
	}

	fmt.Fprintf(&buf, "**URL:** %s\n", card.URL)

	// Description
	if card.Content != "" {
		fmt.Fprint(&buf, "\n## Description\n\n")
		contentMd, err := converter.RichTextToMarkdown(card.Content)
		if err != nil {
			return "", fmt.Errorf("failed to convert card content to markdown: %w", err)
		}
		fmt.Fprintf(&buf, "%s\n", contentMd)
	}

	// Steps
	if len(card.Steps) > 0 {
		fmt.Fprintf(&buf, "\n## Steps (%d)\n\n", len(card.Steps))
		for _, step := range card.Steps {
			checkbox := "[ ]"
			if step.Completed {
				checkbox = "[x]"
			}

			fmt.Fprintf(&buf, "- %s %s", checkbox, step.Title)

			// Add step details
			details := []string{}
			if len(step.Assignees) > 0 {
				assigneeNames := make([]string, len(step.Assignees))
				for i, assignee := range step.Assignees {
					assigneeNames[i] = assignee.Name
				}
				details = append(details, fmt.Sprintf("Assignee: %s", strings.Join(assigneeNames, ", ")))
			}
			if step.DueOn != nil && *step.DueOn != "" {
				details = append(details, fmt.Sprintf("Due: %s", *step.DueOn))
			}

			if len(details) > 0 {
				fmt.Fprintf(&buf, " (%s)", strings.Join(details, ", "))
			}
			fmt.Fprint(&buf, "\n")
		}
	}

	// Comments
	if len(comments) > 0 {
		fmt.Fprintf(&buf, "\n## Comments (%d)\n", len(comments))
		for i, comment := range comments {
			fmt.Fprintf(&buf, "\n### Comment %d - %s (%s)\n\n",
				i+1,
				comment.Creator.Name,
				comment.CreatedAt.Format("Jan 2, 2006 at 3:04 PM"))

			commentMd, err := converter.RichTextToMarkdown(comment.Content)
			if err != nil {
				return "", fmt.Errorf("failed to convert comment content to markdown: %w", err)
			}
			fmt.Fprintf(&buf, "%s\n", commentMd)
		}
	}

	return buf.String(), nil
}

// FormatTodoAsMarkdown formats a todo with all its comments as AI-optimized markdown
func FormatTodoAsMarkdown(todo *api.Todo, comments []api.Comment) (string, error) {
	var buf strings.Builder
	converter := markdown.NewConverter()

	// Title
	fmt.Fprintf(&buf, "# %s\n\n", todo.Title)

	// Metadata
	fmt.Fprintf(&buf, "**ID:** %d\n", todo.ID)
	fmt.Fprintf(&buf, "**Completed:** %t\n", todo.Completed)

	// Parse and format timestamps for consistency with other formatters
	// Todo uses string timestamps while other resources use time.Time
	createdAt, updatedAt := todo.CreatedAt, todo.UpdatedAt
	if parsed, err := parseTime(todo.CreatedAt); err == nil {
		createdAt = parsed.Format("2006-01-02 15:04")
	}
	if parsed, err := parseTime(todo.UpdatedAt); err == nil {
		updatedAt = parsed.Format("2006-01-02 15:04")
	}

	fmt.Fprintf(&buf, "**Created:** %s\n", createdAt)
	fmt.Fprintf(&buf, "**Updated:** %s\n", updatedAt)

	if len(todo.Assignees) > 0 {
		assigneeNames := make([]string, len(todo.Assignees))
		for i, assignee := range todo.Assignees {
			assigneeNames[i] = assignee.Name
		}
		fmt.Fprintf(&buf, "**Assignees:** %s\n", strings.Join(assigneeNames, ", "))
	}

	if todo.DueOn != nil && *todo.DueOn != "" {
		fmt.Fprintf(&buf, "**Due:** %s\n", *todo.DueOn)
	}

	if todo.StartsOn != nil && *todo.StartsOn != "" {
		fmt.Fprintf(&buf, "**Starts:** %s\n", *todo.StartsOn)
	}

	if todo.Creator != nil {
		fmt.Fprintf(&buf, "**Created by:** %s\n", todo.Creator.Name)
	}

	fmt.Fprintf(&buf, "**Todolist ID:** %d\n", todo.TodolistID)

	// Description
	if todo.Description != "" {
		fmt.Fprint(&buf, "\n## Description\n\n")
		descMd, err := converter.RichTextToMarkdown(todo.Description)
		if err != nil {
			return "", fmt.Errorf("failed to convert todo description to markdown: %w", err)
		}
		fmt.Fprintf(&buf, "%s\n", descMd)
	}

	// Comments
	if len(comments) > 0 {
		fmt.Fprintf(&buf, "\n## Comments (%d)\n", len(comments))
		for i, comment := range comments {
			fmt.Fprintf(&buf, "\n### Comment %d - %s (%s)\n\n",
				i+1,
				comment.Creator.Name,
				comment.CreatedAt.Format("Jan 2, 2006 at 3:04 PM"))

			commentMd, err := converter.RichTextToMarkdown(comment.Content)
			if err != nil {
				return "", fmt.Errorf("failed to convert comment content to markdown: %w", err)
			}
			fmt.Fprintf(&buf, "%s\n", commentMd)
		}
	}

	return buf.String(), nil
}

// FormatMessageAsMarkdown formats a message with all its comments as AI-optimized markdown
func FormatMessageAsMarkdown(message *api.Message, comments []api.Comment) (string, error) {
	var buf strings.Builder
	converter := markdown.NewConverter()

	// Title
	fmt.Fprintf(&buf, "# %s\n\n", message.Subject)

	// Metadata
	fmt.Fprintf(&buf, "**ID:** %d\n", message.ID)
	fmt.Fprintf(&buf, "**Status:** %s\n", message.Status)
	fmt.Fprintf(&buf, "**Created:** %s\n", message.CreatedAt.Format("2006-01-02 15:04"))
	fmt.Fprintf(&buf, "**Updated:** %s\n", message.UpdatedAt.Format("2006-01-02 15:04"))

	if message.Creator.Name != "" {
		fmt.Fprintf(&buf, "**Created by:** %s\n", message.Creator.Name)
	}

	if message.Category != nil {
		fmt.Fprintf(&buf, "**Category:** %s\n", message.Category.Name)
	}

	fmt.Fprintf(&buf, "**URL:** %s\n", message.URL)

	// Content
	if message.Content != "" {
		fmt.Fprint(&buf, "\n## Content\n\n")
		contentMd, err := converter.RichTextToMarkdown(message.Content)
		if err != nil {
			return "", fmt.Errorf("failed to convert message content to markdown: %w", err)
		}
		fmt.Fprintf(&buf, "%s\n", contentMd)
	}

	// Comments
	if len(comments) > 0 {
		fmt.Fprintf(&buf, "\n## Comments (%d)\n", len(comments))
		for i, comment := range comments {
			fmt.Fprintf(&buf, "\n### Comment %d - %s (%s)\n\n",
				i+1,
				comment.Creator.Name,
				comment.CreatedAt.Format("Jan 2, 2006 at 3:04 PM"))

			commentMd, err := converter.RichTextToMarkdown(comment.Content)
			if err != nil {
				return "", fmt.Errorf("failed to convert comment content to markdown: %w", err)
			}
			fmt.Fprintf(&buf, "%s\n", commentMd)
		}
	}

	return buf.String(), nil
}

// FormatDocumentAsMarkdown formats a document with all its comments as AI-optimized markdown
func FormatDocumentAsMarkdown(document *api.Document, comments []api.Comment) (string, error) {
	var buf strings.Builder
	converter := markdown.NewConverter()

	// Title
	fmt.Fprintf(&buf, "# %s\n\n", document.Title)

	// Metadata
	fmt.Fprintf(&buf, "**ID:** %d\n", document.ID)
	fmt.Fprintf(&buf, "**Status:** %s\n", document.Status)
	fmt.Fprintf(&buf, "**Created:** %s\n", document.CreatedAt.Format("2006-01-02 15:04"))
	fmt.Fprintf(&buf, "**Updated:** %s\n", document.UpdatedAt.Format("2006-01-02 15:04"))

	if document.Creator.Name != "" {
		fmt.Fprintf(&buf, "**Created by:** %s\n", document.Creator.Name)
	}

	fmt.Fprintf(&buf, "**URL:** %s\n", document.URL)

	// Content
	if document.Content != "" {
		fmt.Fprint(&buf, "\n## Content\n\n")
		contentMd, err := converter.RichTextToMarkdown(document.Content)
		if err != nil {
			return "", fmt.Errorf("failed to convert document content to markdown: %w", err)
		}
		fmt.Fprintf(&buf, "%s\n", contentMd)
	}

	// Comments
	if len(comments) > 0 {
		fmt.Fprintf(&buf, "\n## Comments (%d)\n", len(comments))
		for i, comment := range comments {
			fmt.Fprintf(&buf, "\n### Comment %d - %s (%s)\n\n",
				i+1,
				comment.Creator.Name,
				comment.CreatedAt.Format("Jan 2, 2006 at 3:04 PM"))

			commentMd, err := converter.RichTextToMarkdown(comment.Content)
			if err != nil {
				return "", fmt.Errorf("failed to convert comment content to markdown: %w", err)
			}
			fmt.Fprintf(&buf, "%s\n", commentMd)
		}
	}

	return buf.String(), nil
}
