package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseNextLinkURL(t *testing.T) {
	tests := []struct {
		name       string
		linkHeader string
		expected   string
	}{
		{
			name:       "simple next link",
			linkHeader: `<https://3.basecampapi.com/999999999/buckets/2085958496/messages.json?page=4>; rel="next"`,
			expected:   "https://3.basecampapi.com/999999999/buckets/2085958496/messages.json?page=4",
		},
		{
			name:       "multiple links with next",
			linkHeader: `<https://3.basecampapi.com/999999999/buckets/123/todos.json?page=1>; rel="first", <https://3.basecampapi.com/999999999/buckets/123/todos.json?page=3>; rel="next"`,
			expected:   "https://3.basecampapi.com/999999999/buckets/123/todos.json?page=3",
		},
		{
			name:       "no next link",
			linkHeader: `<https://3.basecampapi.com/999999999/buckets/123/todos.json?page=1>; rel="first"`,
			expected:   "",
		},
		{
			name:       "empty link header",
			linkHeader: "",
			expected:   "",
		},
		{
			name:       "malformed link header",
			linkHeader: `malformed link header`,
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseNextLinkURL(tt.linkHeader)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractPathFromURL(t *testing.T) {
	tests := []struct {
		name        string
		absoluteURL string
		expected    string
	}{
		{
			name:        "basecamp API URL",
			absoluteURL: "https://3.basecampapi.com/999999999/buckets/2085958496/messages.json?page=4",
			expected:    "/buckets/2085958496/messages.json?page=4",
		},
		{
			name:        "todo list URL",
			absoluteURL: "https://3.basecampapi.com/123456789/buckets/987654321/todolists/111/todos.json?page=2",
			expected:    "/buckets/987654321/todolists/111/todos.json?page=2",
		},
		{
			name:        "already relative path",
			absoluteURL: "/buckets/123/todos.json",
			expected:    "/buckets/123/todos.json",
		},
		{
			name:        "malformed URL",
			absoluteURL: "not-a-url",
			expected:    "",
		},
		{
			name:        "empty URL",
			absoluteURL: "",
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractPathFromURL(tt.absoluteURL)
			assert.Equal(t, tt.expected, result)
		})
	}
}