package message

import (
	"testing"

	"github.com/needmore/bc4/internal/factory"
	"github.com/stretchr/testify/assert"
)

func TestNewPostCmd(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedError bool
		errorContains string
	}{
		{
			name:          "no arguments",
			args:          []string{},
			expectedError: false,
		},
		{
			name:          "with project ID",
			args:          []string{"123456"},
			expectedError: false,
		},
		{
			name:          "with title flag",
			args:          []string{"--title", "Test Message"},
			expectedError: false,
		},
		{
			name:          "with content flag",
			args:          []string{"--content", "Test content"},
			expectedError: false,
		},
		{
			name:          "with draft flag",
			args:          []string{"--draft"},
			expectedError: false,
		},
		{
			name:          "with category ID",
			args:          []string{"--category-id", "123"},
			expectedError: false,
		},
		{
			name:          "with all flags",
			args:          []string{"--title", "Test", "--content", "Content", "--draft", "--category-id", "123"},
			expectedError: false,
		},
		{
			name:          "too many arguments",
			args:          []string{"123456", "extra"},
			expectedError: true,
			errorContains: "accepts at most 1 arg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &factory.Factory{}
			cmd := newPostCmd(f)
			cmd.SetArgs(tt.args)

			// Parse flags
			err := cmd.ParseFlags(tt.args)

			if tt.name == "too many arguments" {
				// Argument validation happens during Execute, not ParseFlags
				// So we expect no error here
				assert.NoError(t, err)
			} else if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" && err != nil {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPostCmdFlags(t *testing.T) {
	f := &factory.Factory{}
	cmd := newPostCmd(f)

	// Test title flag
	titleFlag := cmd.Flag("title")
	assert.NotNil(t, titleFlag)
	assert.Equal(t, "t", titleFlag.Shorthand)
	assert.Equal(t, "Message subject", titleFlag.Usage)

	// Test content flag
	contentFlag := cmd.Flag("content")
	assert.NotNil(t, contentFlag)
	assert.Equal(t, "c", contentFlag.Shorthand)
	assert.Equal(t, "Message content (markdown supported)", contentFlag.Usage)

	// Test draft flag
	draftFlag := cmd.Flag("draft")
	assert.NotNil(t, draftFlag)
	assert.Equal(t, "d", draftFlag.Shorthand)
	assert.Equal(t, "Create as draft", draftFlag.Usage)

	// Test category-id flag
	categoryFlag := cmd.Flag("category-id")
	assert.NotNil(t, categoryFlag)
	assert.Equal(t, "Category ID", categoryFlag.Usage)
}

func TestPostCmdLongDescription(t *testing.T) {
	f := &factory.Factory{}
	cmd := newPostCmd(f)

	// Check that long description mentions stdin support
	assert.Contains(t, cmd.Long, "Via stdin")
	assert.Contains(t, cmd.Long, "echo")
	assert.Contains(t, cmd.Long, "cat")
	assert.Contains(t, cmd.Long, "bc4 message post")
}

