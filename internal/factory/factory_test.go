package factory

import (
	"testing"
)

func TestNew(t *testing.T) {
	f := New()
	if f == nil {
		t.Fatal("New() returned nil")
	}

	// Verify initial state
	if f.accountID != "" {
		t.Errorf("Expected empty accountID, got %s", f.accountID)
	}
	if f.projectID != "" {
		t.Errorf("Expected empty projectID, got %s", f.projectID)
	}
}

func TestWithAccount(t *testing.T) {
	f := New()
	accountID := "123456"

	f2 := f.WithAccount(accountID)

	// Should return same instance
	if f != f2 {
		t.Error("WithAccount should return same factory instance")
	}

	// Should set account ID
	if f.accountID != accountID {
		t.Errorf("Expected accountID %s, got %s", accountID, f.accountID)
	}
}

func TestWithProject(t *testing.T) {
	f := New()
	projectID := "789012"

	f2 := f.WithProject(projectID)

	// Should return same instance
	if f != f2 {
		t.Error("WithProject should return same factory instance")
	}

	// Should set project ID
	if f.projectID != projectID {
		t.Errorf("Expected projectID %s, got %s", projectID, f.projectID)
	}
}

func TestConfig_SingleLoad(t *testing.T) {
	// Test that Config() is only loaded once (lazy loading)
	f := New()

	// First call
	cfg1, _ := f.Config()

	// Second call should return same instance
	cfg2, _ := f.Config()

	// If config loading works, both should be the same instance
	if cfg1 != nil && cfg2 != nil && cfg1 != cfg2 {
		t.Error("Config() should return the same instance on multiple calls")
	}
}

func TestAuthClient_ConfigError(t *testing.T) {
	// This test verifies that AuthClient properly handles config errors
	// In a real test environment, we'd mock the config loading
	// For now, we'll just verify the AuthClient method exists and returns appropriate types
	f := New()

	client, err := f.AuthClient()

	// The method should return either a valid client or an error, never both nil
	if client == nil && err == nil {
		t.Error("AuthClient should not return both nil client and nil error")
	}
}

func TestAccountID_Override(t *testing.T) {
	f := New().WithAccount("test-account")

	accountID, err := f.AccountID()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if accountID != "test-account" {
		t.Errorf("Expected accountID 'test-account', got %s", accountID)
	}
}

func TestProjectID_Override(t *testing.T) {
	f := New().WithProject("test-project")

	projectID, err := f.ProjectID()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if projectID != "test-project" {
		t.Errorf("Expected projectID 'test-project', got %s", projectID)
	}
}

func TestContext(t *testing.T) {
	f := New()
	ctx := f.Context()

	if ctx == nil {
		t.Error("Context() returned nil")
	}
}

func TestFactory_LazyLoading(t *testing.T) {
	f := New()

	// Config should not be loaded yet
	if f.config != nil {
		t.Error("Config should not be loaded until requested")
	}

	// AuthClient should not be created yet
	if f.authClient != nil {
		t.Error("AuthClient should not be created until requested")
	}

	// ApiClient should not be created yet
	if f.apiClient != nil {
		t.Error("ApiClient should not be created until requested")
	}
}

func TestApplyOverrides(t *testing.T) {
	tests := []struct {
		name      string
		accountID string
		projectID string
	}{
		{
			name:      "both empty",
			accountID: "",
			projectID: "",
		},
		{
			name:      "account only",
			accountID: "123456",
			projectID: "",
		},
		{
			name:      "project only",
			accountID: "",
			projectID: "789012",
		},
		{
			name:      "both specified",
			accountID: "123456",
			projectID: "789012",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := New() // Create fresh factory for each test
			result := f.ApplyOverrides(tt.accountID, tt.projectID)

			// Should return same factory instance
			if result != f {
				t.Error("ApplyOverrides should return same factory instance")
			}

			// Check that overrides were applied correctly
			if tt.accountID != "" && result.accountID != tt.accountID {
				t.Errorf("Expected accountID %s, got %s", tt.accountID, result.accountID)
			}
			if tt.accountID == "" && result.accountID != "" {
				t.Errorf("Expected empty accountID when not specified, got %s", result.accountID)
			}

			if tt.projectID != "" && result.projectID != tt.projectID {
				t.Errorf("Expected projectID %s, got %s", tt.projectID, result.projectID)
			}
			if tt.projectID == "" && result.projectID != "" {
				t.Errorf("Expected empty projectID when not specified, got %s", result.projectID)
			}
		})
	}
}
