package account

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAuthClientWithError struct {
	mockAuthClient
	setDefaultError error
}

func (m *mockAuthClientWithError) SetDefaultAccount(accountID string) error {
	if m.setDefaultError != nil {
		return m.setDefaultError
	}
	return m.mockAuthClient.SetDefaultAccount(accountID)
}

func TestSetCommand(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		accounts        map[string]auth.AccountToken
		currentDefault  string
		config          *config.Config
		setDefaultError error
		expectOutput    []string
		expectError     bool
		expectDefault   string
		expectConfig    func(*config.Config) // Function to verify config state
	}{
		{
			name:        "no arguments",
			args:        []string{},
			expectError: true,
		},
		{
			name: "set valid account",
			args: []string{"123"},
			accounts: map[string]auth.AccountToken{
				"123": {AccountID: "123", AccountName: "Test Account"},
			},
			expectOutput:  []string{"Default account set to: Test Account (123)"},
			expectDefault: "123",
		},
		{
			name: "account not found",
			args: []string{"999"},
			accounts: map[string]auth.AccountToken{
				"123": {AccountID: "123", AccountName: "Test Account"},
			},
			setDefaultError: fmt.Errorf("account 999 not found"),
			expectError:     true,
		},
		{
			name: "clear default project when changing accounts",
			args: []string{"456"},
			accounts: map[string]auth.AccountToken{
				"123": {AccountID: "123", AccountName: "Account One"},
				"456": {AccountID: "456", AccountName: "Account Two"},
			},
			currentDefault: "123",
			config: &config.Config{
				DefaultProject: "project-123",
			},
			expectOutput:  []string{"Default account set to: Account Two (456)"},
			expectDefault: "456",
			expectConfig: func(c *config.Config) {
				assert.Empty(t, c.DefaultProject, "Default project should be cleared")
			},
		},
		{
			name: "preserve default project when setting same account",
			args: []string{"123"},
			accounts: map[string]auth.AccountToken{
				"123": {AccountID: "123", AccountName: "Test Account"},
			},
			currentDefault: "123",
			config: &config.Config{
				DefaultProject: "project-123",
			},
			expectOutput:  []string{"Default account set to: Test Account (123)"},
			expectDefault: "123",
			expectConfig: func(c *config.Config) {
				assert.Equal(t, "project-123", c.DefaultProject, "Default project should be preserved")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock factory
			var mockAuth factory.AuthClient
			if tt.setDefaultError != nil {
				mockAuth = &mockAuthClientWithError{
					mockAuthClient: mockAuthClient{
						accounts:       tt.accounts,
						defaultAccount: tt.currentDefault,
					},
					setDefaultError: tt.setDefaultError,
				}
			} else {
				mockAuth = &mockAuthClient{
					accounts:       tt.accounts,
					defaultAccount: tt.currentDefault,
				}
			}
			
			// Create config if not provided
			cfg := tt.config
			if cfg == nil {
				cfg = &config.Config{}
			}
			
			fact := &mockFactory{
				authClient: mockAuth,
				config:     cfg,
			}

			// Create command
			cmd := NewCmdSet(fact)
			cmd.SetArgs(tt.args)

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
				
				// Check default account was set
				if tt.expectDefault != "" {
					if mc, ok := mockAuth.(*mockAuthClient); ok {
						assert.Equal(t, tt.expectDefault, mc.defaultAccount)
					}
				}
				
				// Check config state
				if tt.expectConfig != nil {
					tt.expectConfig(cfg)
				}
			}
		})
	}
}

func TestSetCommand_ConfigSave(t *testing.T) {
	// Test that config.Save() is called
	saveCallCount := 0
	
	mockAuth := &mockAuthClient{
		accounts: map[string]auth.AccountToken{
			"123": {AccountID: "123", AccountName: "Test Account"},
		},
	}
	
	cfg := &config.Config{
		DefaultProject: "old-project",
	}
	
	// Override Save method to count calls
	originalSave := cfg.Save
	cfg.Save = func() error {
		saveCallCount++
		return nil
	}
	defer func() { cfg.Save = originalSave }()
	
	fact := &mockFactory{
		authClient: mockAuth,
		config:     cfg,
	}

	cmd := NewCmdSet(fact)
	cmd.SetArgs([]string{"123"})
	
	err := cmd.Execute()
	require.NoError(t, err)
	
	assert.Equal(t, 1, saveCallCount, "Config.Save should be called once")
}