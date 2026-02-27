package mentions

import (
	"context"
	"errors"
	"testing"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/api/mock"
)

func TestResolve(t *testing.T) {
	people := []api.Person{
		{ID: 1, Name: "John Doe", EmailAddress: "john@example.com", AttachableSGID: "sgid-john"},
		{ID: 2, Name: "Jane Smith", EmailAddress: "jane@example.com", AttachableSGID: "sgid-jane"},
		{ID: 3, Name: "Bob Johnson", EmailAddress: "bob@company.com", AttachableSGID: "sgid-bob"},
	}

	tests := []struct {
		name        string
		content     string
		mockPeople  []api.Person
		mockError   error
		expected    string
		expectError bool
	}{
		{
			name:       "no mentions returns content unchanged",
			content:    "<p>Hello world</p>",
			mockPeople: people,
			expected:   "<p>Hello world</p>",
		},
		{
			name:       "single @FirstName mention",
			content:    `<p>Hey @John check this out</p>`,
			mockPeople: people,
			expected:   `<p>Hey <bc-attachment sgid="sgid-john"></bc-attachment> check this out</p>`,
		},
		{
			name:       "single @First.Last mention",
			content:    `<p>Hey @John.Doe check this out</p>`,
			mockPeople: people,
			expected:   `<p>Hey <bc-attachment sgid="sgid-john"></bc-attachment> check this out</p>`,
		},
		{
			name:       "multiple mentions",
			content:    `<p>Hey @John and @Jane</p>`,
			mockPeople: people,
			expected:   `<p>Hey <bc-attachment sgid="sgid-john"></bc-attachment> and <bc-attachment sgid="sgid-jane"></bc-attachment></p>`,
		},
		{
			name:       "mention at start of content",
			content:    `@Bob please review`,
			mockPeople: people,
			expected:   `<bc-attachment sgid="sgid-bob"></bc-attachment> please review`,
		},
		{
			name:       "mention after blockquote marker",
			content:    `<blockquote>@John said something</blockquote>`,
			mockPeople: people,
			expected:   `<blockquote><bc-attachment sgid="sgid-john"></bc-attachment> said something</blockquote>`,
		},
		{
			name:        "unknown mention returns error",
			content:     `<p>Hey @Unknown</p>`,
			mockPeople:  people,
			expectError: true,
		},
		{
			name:        "API error propagated",
			content:     `<p>Hey @John</p>`,
			mockError:   errors.New("API error"),
			expectError: true,
		},
		{
			name:       "empty content returns empty",
			content:    "",
			mockPeople: people,
			expected:   "",
		},
		{
			name:       "content with email address not treated as mention",
			content:    `<p>Email me at john@example.com</p>`,
			mockPeople: people,
			expected:   `<p>Email me at john@example.com</p>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := mock.NewMockClient()
			mockClient.People = tt.mockPeople
			mockClient.PeopleError = tt.mockError

			result, err := Resolve(context.Background(), tt.content, mockClient, "12345")

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected:\n  %s\nGot:\n  %s", tt.expected, result)
			}
		})
	}
}

func TestMentionRegex(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		matches []string
	}{
		{"simple mention", "Hello @John", []string{"@John"}},
		{"dot mention", "Hello @John.Doe", []string{"@John.Doe"}},
		{"start of string", "@John hello", []string{"@John"}},
		{"after blockquote", ">@John said", []string{"@John"}},
		{"multiple mentions", "@John and @Jane", []string{"@John", "@Jane"}},
		{"no mentions", "Hello world", nil},
		{"email not matched", "user@example.com", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			submatches := mentionRe.FindAllStringSubmatch(tt.input, -1)
			var got []string
			for _, sm := range submatches {
				got = append(got, sm[1])
			}

			if len(got) != len(tt.matches) {
				t.Errorf("Expected %d matches, got %d: %v", len(tt.matches), len(got), got)
				return
			}

			for i, m := range got {
				if m != tt.matches[i] {
					t.Errorf("Match[%d]: expected %q, got %q", i, tt.matches[i], m)
				}
			}
		})
	}
}
