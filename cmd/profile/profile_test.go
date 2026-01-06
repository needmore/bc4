package profile

import (
	"testing"

	"github.com/needmore/bc4/internal/factory"
	"github.com/stretchr/testify/assert"
)

func TestNewProfileCmd(t *testing.T) {
	f := factory.New()
	cmd := NewProfileCmd(f)

	// Test basic command properties
	assert.Equal(t, "profile", cmd.Use)
	assert.Equal(t, "Show current user profile", cmd.Short)
	assert.Contains(t, cmd.Long, "Basecamp profile information")
}

func TestProfileCmd_Aliases(t *testing.T) {
	f := factory.New()
	cmd := NewProfileCmd(f)

	// Test that aliases are set correctly
	assert.Contains(t, cmd.Aliases, "me")
	// whoami should NOT be an alias (conflicts with account current)
	assert.NotContains(t, cmd.Aliases, "whoami")
}

func TestProfileCmd_Flags(t *testing.T) {
	f := factory.New()
	cmd := NewProfileCmd(f)

	// Test that json flag exists
	jsonFlag := cmd.Flags().Lookup("json")
	assert.NotNil(t, jsonFlag)
	assert.Equal(t, "false", jsonFlag.DefValue)
}

func TestProfileCmd_ParseFlags(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedError bool
	}{
		{
			name:          "no arguments",
			args:          []string{},
			expectedError: false,
		},
		{
			name:          "json flag",
			args:          []string{"--json"},
			expectedError: false,
		},
		{
			name:          "help flag",
			args:          []string{"--help"},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := factory.New()
			cmd := NewProfileCmd(f)
			cmd.SetArgs(tt.args)

			err := cmd.ParseFlags(tt.args)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func BenchmarkNewProfileCmd(b *testing.B) {
	f := factory.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewProfileCmd(f)
	}
}
