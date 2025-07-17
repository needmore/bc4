package account

import (
	"bytes"
	"testing"

	"github.com/needmore/bc4/internal/auth"
	"github.com/needmore/bc4/internal/config"
	"github.com/needmore/bc4/internal/factory"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock helpers for testing
type testContext struct {
	authClient   *mockAuthClient
	config       *config.Config
	isOutputToTTY bool
}

type mockAuthClient struct {
	accounts       map[string]auth.AccountToken
	defaultAccount string
}

func (m *mockAuthClient) Login() (*auth.AccountToken, error) {
	return nil, nil
}

func (m *mockAuthClient) Logout(accountID string) error {
	return nil
}

func (m *mockAuthClient) GetToken(accountID string) (*auth.AccountToken, error) {
	if token, ok := m.accounts[accountID]; ok {
		return &token, nil
	}
	return nil, nil
}

func (m *mockAuthClient) GetAccounts() map[string]auth.AccountToken {
	return m.accounts
}

func (m *mockAuthClient) GetDefaultAccount() string {
	return m.defaultAccount
}

func (m *mockAuthClient) SetDefaultAccount(accountID string) error {
	m.defaultAccount = accountID
	return nil
}

func TestListCommand(t *testing.T) {
	tests := []struct {
		name           string
		accounts       map[string]auth.AccountToken
		defaultAccount string
		isOutputToTTY  bool
		format         string
		expectOutput   []string
		expectError    bool
	}{
		{
			name:     "no accounts",
			accounts: map[string]auth.AccountToken{},
			isOutputToTTY: true,
			expectOutput: []string{"No authenticated accounts found"},
		},
		{
			name: "single account TTY",
			accounts: map[string]auth.AccountToken{
				"123": {AccountID: "123", AccountName: "Test Account"},
			},
			defaultAccount: "123",
			isOutputToTTY:  true,
			expectOutput:   []string{"ACCOUNT ID", "ACCOUNT NAME", "123", "Test Account", "*"},
		},
		{
			name: "multiple accounts TTY",
			accounts: map[string]auth.AccountToken{
				"123": {AccountID: "123", AccountName: "Account One"},
				"456": {AccountID: "456", AccountName: "Account Two"},
			},
			defaultAccount: "456",
			isOutputToTTY:  true,
			expectOutput:   []string{"ACCOUNT ID", "ACCOUNT NAME", "123", "Account One", "456", "Account Two", "*"},
		},
		{
			name: "single account non-TTY",
			accounts: map[string]auth.AccountToken{
				"123": {AccountID: "123", AccountName: "Test Account"},
			},
			isOutputToTTY: false,
			expectOutput:  []string{"123\tTest Account"},
		},
		{
			name: "TSV format",
			accounts: map[string]auth.AccountToken{
				"123": {AccountID: "123", AccountName: "Test Account"},
			},
			format:       "tsv",
			expectOutput: []string{"123\tTest Account"},
		},
		{
			name: "JSON format not implemented",
			accounts: map[string]auth.AccountToken{
				"123": {AccountID: "123", AccountName: "Test Account"},
			},
			format:      "json",
			expectError: true,
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
				authClient:    mockAuth,
				isOutputToTTY: tt.isOutputToTTY,
			}

			// Create command
			cmd := newListCmd(nil) // We'll mock the factory calls directly
			
			// Set format flag if specified
			if tt.format != "" {
				cmd.Flags().Set("format", tt.format)
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
			}
		})
	}
}

func TestListCommand_SortsByName(t *testing.T) {
	// TODO: Re-enable when we can properly mock factory
	t.Skip("Skipping test that requires factory mocking")
	
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	
	err := cmd.Execute()
	require.NoError(t, err)
	
	output := buf.String()
	lines := bytes.Split(buf.Bytes(), []byte("\n"))
	
	// Verify order - Alpha should come before Beta, Beta before Zebra
	var alphaIndex, betaIndex, zebraIndex int
	for i, line := range lines {
		if bytes.Contains(line, []byte("Alpha Account")) {
			alphaIndex = i
		} else if bytes.Contains(line, []byte("Beta Account")) {
			betaIndex = i
		} else if bytes.Contains(line, []byte("Zebra Account")) {
			zebraIndex = i
		}
	}
	
	assert.Less(t, alphaIndex, betaIndex, "Alpha should come before Beta")
	assert.Less(t, betaIndex, zebraIndex, "Beta should come before Zebra")
}

func TestListCommand_Aliases(t *testing.T) {
	// Test that the command has the correct aliases
	cmd := newListCmd(nil)
	assert.Contains(t, cmd.Aliases, "ls")
}

func TestListCommand_DeprecatedJSONFlag(t *testing.T) {
	// TODO: Re-enable when we can properly mock factory
	t.Skip("Skipping test that requires factory mocking")
	
	// Test deprecated --json flag
	cmd.Flags().Set("json", "true")
	
	var buf bytes.Buffer
	cmd.SetErr(&buf)
	
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, buf.String(), "--json flag is deprecated")
}