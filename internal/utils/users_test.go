package utils

import (
	"context"
	"testing"

	"github.com/needmore/bc4/internal/api"
)

// mockClient is a mock API client for testing
type mockClient struct {
	people []api.Person
}

func (m *mockClient) GetProjectPeople(ctx context.Context, projectID string) ([]api.Person, error) {
	return m.people, nil
}

func TestUserResolver_ResolveUsers(t *testing.T) {
	// Create mock people
	people := []api.Person{
		{ID: 1, Name: "John Doe", EmailAddress: "john@example.com"},
		{ID: 2, Name: "Jane Smith", EmailAddress: "jane@example.com"},
		{ID: 3, Name: "Bob Johnson", EmailAddress: "bob@company.com"},
	}

	// Create mock client
	_ = &mockClient{people: people}

	// Create user resolver with a type assertion (in real code, we'd use an interface)
	// For now, we'll skip the test implementation since it would require refactoring
	// the UserResolver to accept an interface instead of *api.Client
	t.Skip("Test requires refactoring UserResolver to accept an interface")
}

func TestUserResolver_resolveIdentifier(t *testing.T) {
	// Create test people
	ur := &UserResolver{
		people: []api.Person{
			{ID: 1, Name: "John Doe", EmailAddress: "john@example.com"},
			{ID: 2, Name: "Jane Smith", EmailAddress: "jane@example.com"},
			{ID: 3, Name: "Bob Johnson", EmailAddress: "bob@company.com"},
		},
		cached: true,
	}

	tests := []struct {
		name       string
		identifier string
		wantID     int64
		wantFound  bool
	}{
		// Email tests
		{"email exact match", "john@example.com", 1, true},
		{"email case insensitive", "JOHN@EXAMPLE.COM", 1, true},
		{"email with spaces", " jane@example.com ", 2, true},

		// @mention tests
		{"@mention full name", "@John Doe", 1, true},
		{"@mention first name", "@John", 1, true},
		{"@mention last name", "@Doe", 1, true},
		{"@mention case insensitive", "@john", 1, true},
		{"@mention partial match", "@jane", 2, true},

		// Name without @ tests
		{"name without @", "Bob Johnson", 3, true},
		{"first name without @", "Bob", 3, true},

		// Not found tests
		{"unknown email", "unknown@example.com", 0, false},
		{"unknown @mention", "@unknown", 0, false},
		{"empty string", "", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotFound := ur.resolveIdentifier(tt.identifier)
			if gotID != tt.wantID || gotFound != tt.wantFound {
				t.Errorf("resolveIdentifier(%q) = (%d, %v), want (%d, %v)",
					tt.identifier, gotID, gotFound, tt.wantID, tt.wantFound)
			}
		})
	}
}
