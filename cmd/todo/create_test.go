package todo

import (
	"context"
	"errors"
	"testing"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/api/mock"
)

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

func TestCreateTodoCommand(t *testing.T) {
	tests := []struct {
		name            string
		projectID       string
		todoListID      int64
		todoContent     string
		todoDescription string
		assigneeIDs     []int64
		dueOn           *string
		mockTodo        *api.Todo
		mockError       error
		expectError     bool
	}{
		{
			name:        "Create simple todo",
			projectID:   "123",
			todoListID:  456,
			todoContent: "Write unit tests",
			mockTodo: &api.Todo{
				ID:      789,
				Content: "Write unit tests",
			},
			expectError: false,
		},
		{
			name:            "Create todo with all fields",
			projectID:       "123",
			todoListID:      456,
			todoContent:     "Complete feature",
			todoDescription: "Implement the new feature with tests",
			assigneeIDs:     []int64{101, 102},
			dueOn:           stringPtr("2024-12-31"),
			mockTodo: &api.Todo{
				ID:          789,
				Content:     "Complete feature",
				Description: "Implement the new feature with tests",
			},
			expectError: false,
		},
		{
			name:        "API error",
			projectID:   "123",
			todoListID:  456,
			todoContent: "Failed todo",
			mockError:   errors.New("API error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := mock.NewMockClient()
			mockClient.CreatedTodo = tt.mockTodo
			mockClient.CreateTodoError = tt.mockError

			// Create the request
			req := api.TodoCreateRequest{
				Content:     tt.todoContent,
				Description: tt.todoDescription,
				AssigneeIDs: tt.assigneeIDs,
				DueOn:       tt.dueOn,
			}

			// Call the API
			ctx := context.Background()
			todo, err := mockClient.CreateTodo(ctx, tt.projectID, tt.todoListID, req)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify the returned todo
			if !tt.expectError && todo != nil {
				if todo.Content != tt.todoContent {
					t.Errorf("Expected todo content %s, got %s", tt.todoContent, todo.Content)
				}
			}

			// Verify API was called
			if len(mockClient.Calls) != 1 {
				t.Errorf("Expected 1 API call, got %d", len(mockClient.Calls))
			}
		})
	}
}

func TestCompleteTodoCommand(t *testing.T) {
	tests := []struct {
		name        string
		projectID   string
		todoID      int64
		mockError   error
		expectError bool
	}{
		{
			name:        "Complete todo successfully",
			projectID:   "123",
			todoID:      789,
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "API error",
			projectID:   "123",
			todoID:      789,
			mockError:   errors.New("API error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := mock.NewMockClient()
			mockClient.CompleteTodoError = tt.mockError

			// Call the API
			ctx := context.Background()
			err := mockClient.CompleteTodo(ctx, tt.projectID, tt.todoID)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify API was called
			expectedCall := "CompleteTodo(123, 789)"
			if len(mockClient.Calls) != 1 || mockClient.Calls[0] != expectedCall {
				t.Errorf("Expected call %s, got %v", expectedCall, mockClient.Calls)
			}
		})
	}
}

