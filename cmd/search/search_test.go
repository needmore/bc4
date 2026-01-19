package search

import (
	"strconv"
	"testing"

	"github.com/needmore/bc4/internal/factory"
	"github.com/stretchr/testify/assert"
)

func TestNewSearchCmd(t *testing.T) {
	// Create factory
	f := factory.New()

	// Create search command
	cmd := NewSearchCmd(f)

	// Test basic properties
	assert.Equal(t, "search <query>", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Test that it requires exactly 1 argument
	assert.NotNil(t, cmd.Args)

	// Test that flags are defined
	flags := cmd.Flags()
	assert.NotNil(t, flags.Lookup("type"))
	assert.NotNil(t, flags.Lookup("project"))
	assert.NotNil(t, flags.Lookup("account"))
	assert.NotNil(t, flags.Lookup("format"))
	assert.NotNil(t, flags.Lookup("limit"))

	// Test default values
	limit, err := flags.GetInt("limit")
	assert.NoError(t, err)
	assert.Equal(t, 50, limit)

	format, err := flags.GetString("format")
	assert.NoError(t, err)
	assert.Equal(t, "table", format)
}

func TestParseResourceTypes(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    []string
		shouldError bool
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
			name:     "types with spaces",
			input:    "todo, message , document",
			expected: []string{"Todo", "Message", "Document"},
		},
		{
			name:     "case insensitive",
			input:    "TODO,Message,DOCUMENT",
			expected: []string{"Todo", "Message", "Document"},
		},
		{
			name:     "card type",
			input:    "card",
			expected: []string{"Card"},
		},
		{
			name:        "invalid type",
			input:       "invalid",
			shouldError: true,
		},
		{
			name:        "mixed valid and invalid",
			input:       "todo,invalid",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseResourceTypes(tt.input)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestSearchCommand_LimitValidation(t *testing.T) {
	f := factory.New()
	cmd := NewSearchCmd(f)

	tests := []struct {
		name        string
		limit       int
		shouldError bool
	}{
		{
			name:  "valid limit",
			limit: 50,
		},
		{
			name:  "zero limit (no limit)",
			limit: 0,
		},
		{
			name:  "max limit",
			limit: maxSearchLimit,
		},
		{
			name:        "negative limit",
			limit:       -1,
			shouldError: true,
		},
		{
			name:        "exceeds max limit",
			limit:       maxSearchLimit + 1,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the flag
			err := cmd.Flags().Set("limit", strconv.Itoa(tt.limit))
			assert.NoError(t, err)

			// The validation happens in RunE, so we would need to actually
			// run the command to test it. For now, we're just testing
			// that the flag accepts the values. Full integration testing
			// would require mocking the API client.
		})
	}
}

func TestSearchCommand_Flags(t *testing.T) {
	f := factory.New()
	cmd := NewSearchCmd(f)

	// Test that all required flags exist
	requiredFlags := []string{"type", "project", "account", "format", "limit"}
	for _, flag := range requiredFlags {
		assert.NotNil(t, cmd.Flags().Lookup(flag), "Flag %s should exist", flag)
	}

	// Test flag shortcuts
	assert.NotNil(t, cmd.Flags().ShorthandLookup("t"), "Type flag should have shorthand -t")
	assert.NotNil(t, cmd.Flags().ShorthandLookup("p"), "Project flag should have shorthand -p")
	assert.NotNil(t, cmd.Flags().ShorthandLookup("a"), "Account flag should have shorthand -a")
	assert.NotNil(t, cmd.Flags().ShorthandLookup("f"), "Format flag should have shorthand -f")
	assert.NotNil(t, cmd.Flags().ShorthandLookup("l"), "Limit flag should have shorthand -l")
}
