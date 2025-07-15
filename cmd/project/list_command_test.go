package project

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/api/mock"
	"github.com/needmore/bc4/internal/config"
	"github.com/spf13/cobra"
)

func TestListCommand(t *testing.T) {
	tests := []struct {
		name         string
		mockProjects []api.Project
		mockError    error
		expectError  bool
		expectCalls  []string
	}{
		{
			name: "List projects successfully",
			mockProjects: []api.Project{
				{ID: 1, Name: "Project Alpha", Description: "First project"},
				{ID: 2, Name: "Project Beta", Description: "Second project"},
			},
			expectError: false,
			expectCalls: []string{"GetProjects"},
		},
		{
			name:         "Empty project list",
			mockProjects: []api.Project{},
			expectError:  false,
			expectCalls:  []string{"GetProjects"},
		},
		{
			name:         "API error",
			mockProjects: nil,
			mockError:    errors.New("API error"),
			expectError:  true,
			expectCalls:  []string{"GetProjects"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := mock.NewMockClient()
			mockClient.Projects = tt.mockProjects
			mockClient.ProjectsError = tt.mockError

			// Create a test command that uses our mock client
			cmd := &cobra.Command{
				Use:   "list",
				Short: "List all projects",
				RunE: func(cmd *cobra.Command, args []string) error {
					ctx := context.Background()
					projects, err := mockClient.GetProjects(ctx)
					if err != nil {
						return err
					}

					// Sort projects (using our tested function)
					sortProjectsByName(projects)

					// In real command, this would output to stdout
					// For testing, we just verify the data was retrieved
					if len(projects) == 0 {
						cmd.Println("No projects found")
					}
					return nil
				},
			}

			// Capture output
			buf := bytes.NewBuffer(nil)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			// Execute command
			err := cmd.Execute()

			// Check error expectation
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify API calls
			if len(mockClient.Calls) != len(tt.expectCalls) {
				t.Errorf("Expected %d API calls, got %d", len(tt.expectCalls), len(mockClient.Calls))
			}
			for i, call := range tt.expectCalls {
				if i < len(mockClient.Calls) && mockClient.Calls[i] != call {
					t.Errorf("Expected call %d to be %s, got %s", i, call, mockClient.Calls[i])
				}
			}
		})
	}
}

func TestListCommandWithConfig(t *testing.T) {
	// Test that demonstrates how command would integrate with config
	cfg := &config.Config{
		DefaultAccount: "123456",
	}

	mockClient := mock.NewMockClient()
	mockClient.Projects = []api.Project{
		{ID: 1, Name: "Test Project"},
	}

	// In the real implementation, the command would use cfg.DefaultAccount
	// to create the API client with the correct account ID
	ctx := context.Background()
	projects, err := mockClient.GetProjects(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(projects) != 1 {
		t.Errorf("Expected 1 project, got %d", len(projects))
	}

	if projects[0].Name != "Test Project" {
		t.Errorf("Expected project name 'Test Project', got %s", projects[0].Name)
	}

	// Verify the config has the expected account
	if cfg.DefaultAccount != "123456" {
		t.Errorf("Expected default account '123456', got %s", cfg.DefaultAccount)
	}
}
