package message

import (
	"testing"

	"github.com/needmore/bc4/internal/factory"
	"github.com/stretchr/testify/assert"
)

func TestNewPinCmd(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedError bool
		errorContains string
	}{
		{
			name:          "no arguments",
			args:          []string{},
			expectedError: true,
			errorContains: "message-id",
		},
		{
			name:          "with message ID",
			args:          []string{"123456"},
			expectedError: false,
		},
		{
			name:          "with URL",
			args:          []string{"https://3.basecamp.com/123/buckets/456/messages/789"},
			expectedError: false,
		},
		{
			name:          "too many arguments",
			args:          []string{"123456", "extra"},
			expectedError: true,
			errorContains: "message-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &factory.Factory{}
			cmd := newPinCmd(f)
			cmd.SetArgs(tt.args)

			err := cmd.ValidateArgs(tt.args)

			if tt.expectedError {
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

func TestPinCmdProperties(t *testing.T) {
	f := &factory.Factory{}
	cmd := newPinCmd(f)

	assert.Equal(t, "pin <message-id|url>", cmd.Use)
	assert.Equal(t, "Pin a message to the top of the message board", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotEmpty(t, cmd.Example)
}
