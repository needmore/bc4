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
		{
			name:       "unquoted rel parameter",
			linkHeader: `<https://api.example.com/items?page=2>; rel=next`,
			expected:   "https://api.example.com/items?page=2",
		},
		{
			name:       "quoted parameter with comma inside",
			linkHeader: `<https://api.example.com/search?q=hello,world>; title="Items, Page 2"; rel="next"`,
			expected:   "https://api.example.com/search?q=hello,world",
		},
		{
			name:       "multiple rels in one parameter",
			linkHeader: `<https://api.example.com/page2>; rel="next last"`,
			expected:   "https://api.example.com/page2",
		},
		{
			name:       "case insensitive rel matching",
			linkHeader: `<https://api.example.com/page2>; rel="NEXT"`,
			expected:   "https://api.example.com/page2",
		},
		{
			name:       "complex real-world example",
			linkHeader: `<https://3.basecampapi.com/999999999/buckets/123/todos.json?page=1>; rel="first", <https://3.basecampapi.com/999999999/buckets/123/todos.json?page=2>; rel="prev", <https://3.basecampapi.com/999999999/buckets/123/todos.json?page=4>; rel="next", <https://3.basecampapi.com/999999999/buckets/123/todos.json?page=10>; rel="last"`,
			expected:   "https://3.basecampapi.com/999999999/buckets/123/todos.json?page=4",
		},
		{
			name:       "extra whitespace handling",
			linkHeader: `  <https://api.example.com/page2>  ;  rel="next"  ,  <https://api.example.com/page3>  ;  rel="last"  `,
			expected:   "https://api.example.com/page2",
		},
		{
			name:       "no angle brackets (malformed but graceful)",
			linkHeader: `rel="next"`,
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

func TestParseLinkHeaderEntries(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected []LinkEntry
	}{
		{
			name:   "single link",
			header: `<https://api.example.com/page2>; rel="next"`,
			expected: []LinkEntry{
				{
					URL: "https://api.example.com/page2",
					Params: map[string]string{
						"rel": "next",
					},
				},
			},
		},
		{
			name:   "multiple parameters",
			header: `<https://api.example.com/page2>; rel="next"; title="Next Page"; type="text/html"`,
			expected: []LinkEntry{
				{
					URL: "https://api.example.com/page2",
					Params: map[string]string{
						"rel":   "next",
						"title": "Next Page",
						"type":  "text/html",
					},
				},
			},
		},
		{
			name:   "quoted parameter with comma",
			header: `<https://api.example.com/search?q=hello,world>; title="Items, Page 2"`,
			expected: []LinkEntry{
				{
					URL: "https://api.example.com/search?q=hello,world",
					Params: map[string]string{
						"title": "Items, Page 2",
					},
				},
			},
		},
		{
			name:   "multiple links",
			header: `<https://api.example.com/page1>; rel="first", <https://api.example.com/page3>; rel="next"`,
			expected: []LinkEntry{
				{
					URL: "https://api.example.com/page1",
					Params: map[string]string{
						"rel": "first",
					},
				},
				{
					URL: "https://api.example.com/page3",
					Params: map[string]string{
						"rel": "next",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLinkHeaderEntries(tt.header)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLinkEntry_HasRelation(t *testing.T) {
	tests := []struct {
		name     string
		entry    LinkEntry
		relation string
		expected bool
	}{
		{
			name: "single quoted relation",
			entry: LinkEntry{
				Params: map[string]string{"rel": "next"},
			},
			relation: "next",
			expected: true,
		},
		{
			name: "multiple relations",
			entry: LinkEntry{
				Params: map[string]string{"rel": "next last"},
			},
			relation: "next",
			expected: true,
		},
		{
			name: "case insensitive",
			entry: LinkEntry{
				Params: map[string]string{"rel": "NEXT"},
			},
			relation: "next",
			expected: true,
		},
		{
			name: "no matching relation",
			entry: LinkEntry{
				Params: map[string]string{"rel": "first"},
			},
			relation: "next",
			expected: false,
		},
		{
			name: "no rel parameter",
			entry: LinkEntry{
				Params: map[string]string{"title": "test"},
			},
			relation: "next",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.entry.hasRelation(tt.relation)
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