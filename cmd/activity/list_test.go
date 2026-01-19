package activity

import (
	"testing"
	"time"
)

func TestParseSince(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "hours",
			input:   "24h",
			wantErr: false,
		},
		{
			name:    "days",
			input:   "7d",
			wantErr: false,
		},
		{
			name:    "weeks",
			input:   "2w",
			wantErr: false,
		},
		{
			name:    "RFC3339",
			input:   "2024-01-01T00:00:00Z",
			wantErr: false,
		},
		{
			name:    "date only",
			input:   "2024-01-01",
			wantErr: false,
		},
		{
			name:    "today",
			input:   "today",
			wantErr: false,
		},
		{
			name:    "yesterday",
			input:   "yesterday",
			wantErr: false,
		},
		{
			name:    "this week",
			input:   "this week",
			wantErr: false,
		},
		{
			name:    "last week",
			input:   "last week",
			wantErr: false,
		},
		{
			name:    "invalid",
			input:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseSince(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseSince() expected error for input %q, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("parseSince() unexpected error for input %q: %v", tt.input, err)
				return
			}

			// Verify the result is before now
			if !result.Before(now) {
				t.Errorf("parseSince() result should be before now for input %q", tt.input)
			}
		})
	}
}

func TestParseTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single type",
			input:    "todo",
			expected: []string{"Todo"},
		},
		{
			name:     "multiple types",
			input:    "todo,message,document",
			expected: []string{"Todo", "Message", "Document"},
		},
		{
			name:     "aliases",
			input:    "todos,msg,doc",
			expected: []string{"Todo", "Message", "Document"},
		},
		{
			name:     "mixed case",
			input:    "TODO,Message",
			expected: []string{"Todo", "Message"},
		},
		{
			name:     "with spaces",
			input:    "todo , message , document",
			expected: []string{"Todo", "Message", "Document"},
		},
		{
			name:     "file alias",
			input:    "file,files",
			expected: []string{"Upload", "Upload"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTypes(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("parseTypes() got %d types, expected %d", len(result), len(tt.expected))
				return
			}

			for i, r := range result {
				if r != tt.expected[i] {
					t.Errorf("parseTypes() result[%d] = %q, expected %q", i, r, tt.expected[i])
				}
			}
		})
	}
}

func TestFormatRecordingType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Todo",
			input:    "Todo",
			expected: "todo",
		},
		{
			name:     "Message",
			input:    "Message",
			expected: "message",
		},
		{
			name:     "Document",
			input:    "Document",
			expected: "document",
		},
		{
			name:     "Comment",
			input:    "Comment",
			expected: "comment",
		},
		{
			name:     "Upload",
			input:    "Upload",
			expected: "upload",
		},
		{
			name:     "Schedule Entry",
			input:    "Schedule::Entry",
			expected: "event",
		},
		{
			name:     "Card",
			input:    "Card",
			expected: "card",
		},
		{
			name:     "Card Table",
			input:    "Card::Table",
			expected: "card_table",
		},
		{
			name:     "Unknown type",
			input:    "UnknownType",
			expected: "unknowntype",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatRecordingType(tt.input)
			if result != tt.expected {
				t.Errorf("formatRecordingType(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseDurationValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
		wantErr  bool
	}{
		{
			name:     "single digit",
			input:    "7",
			expected: 7,
			wantErr:  false,
		},
		{
			name:     "multiple digits",
			input:    "24",
			expected: 24,
			wantErr:  false,
		},
		{
			name:     "invalid",
			input:    "abc",
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDurationValue(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseDurationValue() expected error for input %q, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("parseDurationValue() unexpected error for input %q: %v", tt.input, err)
				return
			}

			if result != tt.expected {
				t.Errorf("parseDurationValue(%q) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}
