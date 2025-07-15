package utils

import (
	"context"
	"errors"
	"testing"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/api/mock"
)

func TestNewUserResolver(t *testing.T) {
	mockClient := mock.NewMockClient()
	projectID := "12345"

	resolver := NewUserResolver(mockClient, projectID)

	if resolver == nil {
		t.Fatal("Expected non-nil resolver")
	}

	if resolver.client != mockClient {
		t.Error("Expected client to be set correctly")
	}

	if resolver.projectID != projectID {
		t.Errorf("Expected projectID %s, got %s", projectID, resolver.projectID)
	}

	if resolver.cached {
		t.Error("Expected cached to be false initially")
	}
}

func TestUserResolver_GetPeople(t *testing.T) {
	tests := []struct {
		name        string
		mockPeople  []api.Person
		mockError   error
		expectError bool
		expectCount int
	}{
		{
			name: "Get people successfully",
			mockPeople: []api.Person{
				{ID: 1, Name: "John Doe", EmailAddress: "john@example.com"},
				{ID: 2, Name: "Jane Smith", EmailAddress: "jane@example.com"},
			},
			expectError: false,
			expectCount: 2,
		},
		{
			name:        "Empty people list",
			mockPeople:  []api.Person{},
			expectError: false,
			expectCount: 0,
		},
		{
			name:        "API error",
			mockError:   errors.New("API error"),
			expectError: true,
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := mock.NewMockClient()
			mockClient.People = tt.mockPeople
			mockClient.PeopleError = tt.mockError

			resolver := NewUserResolver(mockClient, "12345")
			ctx := context.Background()

			people, err := resolver.GetPeople(ctx)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if len(people) != tt.expectCount {
				t.Errorf("Expected %d people, got %d", tt.expectCount, len(people))
			}

			// Verify caching behavior
			if !tt.expectError && !resolver.cached {
				t.Error("Expected people to be cached after successful fetch")
			}
		})
	}
}

func TestUserResolver_ResolveUsers(t *testing.T) {
	// Create mock people
	people := []api.Person{
		{ID: 1, Name: "John Doe", EmailAddress: "john@example.com"},
		{ID: 2, Name: "Jane Smith", EmailAddress: "jane@example.com"},
		{ID: 3, Name: "Bob Johnson", EmailAddress: "bob@company.com"},
	}

	tests := []struct {
		name        string
		identifiers []string
		mockPeople  []api.Person
		mockError   error
		expectIDs   []int64
		expectError bool
	}{
		{
			name:        "Resolve single email",
			identifiers: []string{"john@example.com"},
			mockPeople:  people,
			expectIDs:   []int64{1},
			expectError: false,
		},
		{
			name:        "Resolve multiple emails",
			identifiers: []string{"john@example.com", "jane@example.com"},
			mockPeople:  people,
			expectIDs:   []int64{1, 2},
			expectError: false,
		},
		{
			name:        "Resolve @mentions",
			identifiers: []string{"@John", "@Jane"},
			mockPeople:  people,
			expectIDs:   []int64{1, 2},
			expectError: false,
		},
		{
			name:        "Resolve mixed identifiers",
			identifiers: []string{"@Bob", "jane@example.com"},
			mockPeople:  people,
			expectIDs:   []int64{3, 2},
			expectError: false,
		},
		{
			name:        "Skip empty strings",
			identifiers: []string{"", "john@example.com", "  "},
			mockPeople:  people,
			expectIDs:   []int64{1},
			expectError: false,
		},
		{
			name:        "Avoid duplicates",
			identifiers: []string{"john@example.com", "@John", "John Doe"},
			mockPeople:  people,
			expectIDs:   []int64{1},
			expectError: false,
		},
		{
			name:        "Error on unknown users",
			identifiers: []string{"unknown@example.com", "@Unknown"},
			mockPeople:  people,
			expectIDs:   nil,
			expectError: true,
		},
		{
			name:        "Error on API failure",
			identifiers: []string{"john@example.com"},
			mockError:   errors.New("API error"),
			expectIDs:   nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := mock.NewMockClient()
			mockClient.People = tt.mockPeople
			mockClient.PeopleError = tt.mockError

			resolver := NewUserResolver(mockClient, "12345")
			ctx := context.Background()

			ids, err := resolver.ResolveUsers(ctx, tt.identifiers)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				if len(ids) != len(tt.expectIDs) {
					t.Errorf("Expected %d IDs, got %d", len(tt.expectIDs), len(ids))
				}
				for i, id := range ids {
					if i < len(tt.expectIDs) && id != tt.expectIDs[i] {
						t.Errorf("Expected ID[%d] = %d, got %d", i, tt.expectIDs[i], id)
					}
				}
			}
		})
	}
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

func TestWriteToPager(t *testing.T) {
	tests := []struct {
		name        string
		options     *PagerOptions
		expectError bool
	}{
		{
			name:        "nil options",
			options:     nil,
			expectError: false,
		},
		{
			name: "empty options",
			options: &PagerOptions{
				Pager:   "",
				Force:   false,
				NoPager: false,
			},
			expectError: false,
		},
		{
			name: "with pager command",
			options: &PagerOptions{
				Pager:   "less -R",
				Force:   false,
				NoPager: false,
			},
			expectError: false,
		},
		{
			name: "no pager option",
			options: &PagerOptions{
				Pager:   "",
				Force:   false,
				NoPager: true,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer, err := WriteToPager(tt.options)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if writer != nil {
				// Test writing some data
				_, writeErr := writer.Write([]byte("test data\n"))
				if writeErr != nil {
					t.Errorf("Error writing to pager: %v", writeErr)
				}

				// Close the writer
				closeErr := writer.Close()
				if closeErr != nil {
					t.Errorf("Error closing pager: %v", closeErr)
				}
			}
		})
	}
}

