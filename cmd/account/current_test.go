package account

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCurrentCommand(t *testing.T) {
	tests := []struct {
		name           string
		accounts       map[string]auth.AccountToken
		defaultAccount string
		config         *config.Config
		jsonOutput     bool
		expectOutput   []string
		expectError    bool
	}{
		{
			name:         "no default account",
			accounts:     map[string]auth.AccountToken{},
			expectOutput: []string{"No default account set"},
		},
		{
			name: "default account with no info",
			accounts: map[string]auth.AccountToken{
				"999": {AccountID: "999", AccountName: "Other Account"},
			},
			defaultAccount: "123", // Default account not in accounts list
			expectOutput:   []string{"Default account ID: 123", "(Account information not available)"},
		},
		{
			name: "valid default account",
			accounts: map[string]auth.AccountToken{
				"123": {AccountID: "123", AccountName: "Test Account"},
			},
			defaultAccount: "123",
			expectOutput:   []string{"Account ID: 123", "Account Name: Test Account"},
		},
		{
			name: "with global default project",
			accounts: map[string]auth.AccountToken{
				"123": {AccountID: "123", AccountName: "Test Account"},
			},
			defaultAccount: "123",
			config: &config.Config{
				DefaultProject: "456",
			},
			expectOutput: []string{"Account ID: 123", "Account Name: Test Account", "Default Project: 456"},
		},
		{
			name: "with account-specific default project",
			accounts: map[string]auth.AccountToken{
				"123": {AccountID: "123", AccountName: "Test Account"},
			},
			defaultAccount: "123",
			config: &config.Config{
				DefaultProject: "456",
				Accounts: map[string]config.AccountConfig{
					"123": {DefaultProject: "789"},
				},
			},
			expectOutput: []string{"Account ID: 123", "Account Name: Test Account", "Default Project: 789"},
		},
		{
			name: "JSON output",
			accounts: map[string]auth.AccountToken{
				"123": {AccountID: "123", AccountName: "Test Account"},
			},
			defaultAccount: "123",
			jsonOutput:     true,
			expectOutput:   []string{`"account_id":"123"`, `"account_name":"Test Account"`},
		},
		{
			name: "JSON output with default project",
			accounts: map[string]auth.AccountToken{
				"123": {AccountID: "123", AccountName: "Test Account"},
			},
			defaultAccount: "123",
			config: &config.Config{
				DefaultProject: "456",
			},
			jsonOutput:   true,
			expectOutput: []string{`"account_id":"123"`, `"account_name":"Test Account"`, `"default_project":"456"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock factory
			mockAuth := &mockAuthClient{
				accounts:       tt.accounts,
				defaultAccount: tt.defaultAccount,
			}
			fact := &mockFactory{
				authClient: mockAuth,
				config:     tt.config,
			}

			// Create command
			cmd := NewCmdCurrent(fact)
			
			// Set JSON flag if specified
			if tt.jsonOutput {
				cmd.Flags().Set("json", "true")
			}

			// Capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Execute command
			err := cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				output := buf.String()
				
				// Check expected output
				for _, expected := range tt.expectOutput {
					assert.Contains(t, output, expected)
				}
				
				// If JSON output, verify it's valid JSON
				if tt.jsonOutput {
					var result map[string]interface{}
					err := json.Unmarshal(buf.Bytes(), &result)
					assert.NoError(t, err, "Output should be valid JSON")
				}
			}
		})
	}
}

func TestCurrentCommand_Aliases(t *testing.T) {
	// Test that the command has the correct aliases
	fact := &mockFactory{
		authClient: &mockAuthClient{},
	}
	
	cmd := NewCmdCurrent(fact)
	assert.Contains(t, cmd.Aliases, "whoami")
}

func TestCurrentCommand_JSONStructure(t *testing.T) {
	// Test the exact JSON structure
	mockAuth := &mockAuthClient{
		accounts: map[string]auth.AccountToken{
			"123": {AccountID: "123", AccountName: "Test Account"},
		},
		defaultAccount: "123",
	}
	fact := &mockFactory{
		authClient: mockAuth,
		config: &config.Config{
			DefaultProject: "456",
		},
	}

	cmd := NewCmdCurrent(fact)
	cmd.Flags().Set("json", "true")
	
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	
	err := cmd.Execute()
	require.NoError(t, err)
	
	var result map[string]string
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	
	assert.Equal(t, "123", result["account_id"])
	assert.Equal(t, "Test Account", result["account_name"])
	assert.Equal(t, "456", result["default_project"])
}