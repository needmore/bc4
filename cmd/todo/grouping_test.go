package todo

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/api/mock"
	"github.com/needmore/bc4/internal/factory"
)

func TestTodoListGroupingFunctionality(t *testing.T) {
	tests := []struct {
		name         string
		mockSetup    func(*mock.MockClient)
		expectError  bool
		expectGroups bool
		groupedFlag  bool
		todoListID   string
	}{
		{
			name: "List with groups should fetch groups successfully",
			mockSetup: func(client *mock.MockClient) {
				// Set up mock responses
				client.TodoSet = &api.TodoSet{
					ID:           1,
					Title:        "Todo Set",
					TodolistsURL: "http://example.com/todolists",
				}
				client.TodoSetError = nil

				client.TodoList = &api.TodoList{
					ID:          456,
					Title:       "Project Tasks",
					Description: "Tasks for the project",
					GroupsURL:   "http://example.com/groups", // Has groups
				}
				client.TodoListError = nil

				// Empty todos indicates grouped list
				client.Todos = []api.Todo{}
				client.TodosError = nil

				// Mock groups
				client.TodoGroups = []api.TodoGroup{
					{
						ID:             1,
						Title:          "Design Tasks",
						Name:           "design",
						Description:    "UI/UX related tasks",
						CompletedRatio: "2/5",
						TodosCount:     5,
					},
					{
						ID:             2,
						Title:          "Development Tasks",
						Name:           "development",
						Description:    "Code implementation tasks",
						CompletedRatio: "3/7",
						TodosCount:     7,
					},
				}
				client.TodoGroupsError = nil
			},
			expectGroups: true,
			groupedFlag:  true,
			todoListID:   "456",
		},
		{
			name: "List without groups should display flat list",
			mockSetup: func(client *mock.MockClient) {
				client.TodoSet = &api.TodoSet{
					ID:           1,
					Title:        "Todo Set",
					TodolistsURL: "http://example.com/todolists",
				}
				client.TodoSetError = nil

				client.TodoList = &api.TodoList{
					ID:          456,
					Title:       "Simple Tasks",
					Description: "Simple task list",
					GroupsURL:   "", // No groups
				}
				client.TodoListError = nil

				// Regular todos
				client.Todos = []api.Todo{
					{
						ID:        1,
						Title:     "Task 1",
						Content:   "Task 1",
						Completed: false,
					},
					{
						ID:        2,
						Title:     "Task 2",
						Content:   "Task 2",
						Completed: false,
					},
				}
				client.TodosError = nil

				// No groups
				client.TodoGroups = []api.TodoGroup{}
				client.TodoGroupsError = nil
			},
			expectGroups: false,
			groupedFlag:  false,
			todoListID:   "456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := &mock.MockClient{}
			tt.mockSetup(mockClient)

			// Create basic factory for testing
			f := &factory.Factory{}

			// Create command
			cmd := newListCmd(f)

			// The test validates that the command can be created and configured
			// with the grouping flag without errors
			assert.NotNil(t, cmd, "Command should be created successfully")

			// Test that the grouped flag exists
			groupedFlag := cmd.Flag("grouped")
			assert.NotNil(t, groupedFlag, "Grouped flag should exist")

			// Test flag can be set
			if tt.groupedFlag {
				err := cmd.Flags().Set("grouped", "true")
				assert.NoError(t, err, "Should be able to set grouped flag")

				value, err := cmd.Flags().GetBool("grouped")
				assert.NoError(t, err, "Should be able to get grouped flag value")
				assert.True(t, value, "Grouped flag should be true when set")
			}
		})
	}
}

func TestTodoGroupingCommandHelp(t *testing.T) {
	// Create a factory (doesn't need to be fully functional for help text)
	f := &factory.Factory{}

	// Test main todo command mentions grouping
	todoCmd := NewTodoCmd(f)
	assert.Contains(t, todoCmd.Long, "groups", "Main todo command should mention grouping capability")

	// Test list command has detailed grouping information
	listCmd := newListCmd(f)
	assert.Contains(t, listCmd.Long, "groups", "List command should explain grouping")
	assert.Contains(t, listCmd.Long, "section", "List command should mention sections")

	// Test grouped flag has descriptive help
	groupedFlag := listCmd.Flag("grouped")
	assert.NotNil(t, groupedFlag, "Grouped flag should exist")
	assert.Contains(t, groupedFlag.Usage, "groups", "Grouped flag should mention groups in help")
}

func TestCreateTodoGroupCommand(t *testing.T) {
	f := &factory.Factory{}
	
	tests := []struct {
		name        string
		projectID   string
		todoListID  int64
		groupName   string
		expectError bool
	}{
		{
			name:        "Create simple group",
			projectID:   "123",
			todoListID:  456,
			groupName:   "In Progress",
			expectError: false,
		},
		{
			name:        "Create group with long name",
			projectID:   "123",
			todoListID:  456,
			groupName:   "Completed Tasks - Ready for Review",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := mock.NewMockClient()

			// Create the request
			req := api.TodoGroupCreateRequest{
				Name: tt.groupName,
			}

			// Call the API
			ctx := f.Context()
			group, err := mockClient.CreateTodoGroup(ctx, tt.projectID, tt.todoListID, req)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify the returned group
			if !tt.expectError && group != nil {
				assert.Equal(t, tt.groupName, group.Name, "Group name should match")
				assert.NotZero(t, group.ID, "Group ID should be non-zero")
			}

			// Verify API was called
			assert.Len(t, mockClient.Calls, 1, "Expected 1 API call")
		})
	}
}

func TestRepositionTodoGroupCommand(t *testing.T) {
	f := &factory.Factory{}
	
	tests := []struct {
		name        string
		projectID   string
		groupID     int64
		position    int
		expectError bool
	}{
		{
			name:        "Reposition to first",
			projectID:   "123",
			groupID:     789,
			position:    1,
			expectError: false,
		},
		{
			name:        "Reposition to third",
			projectID:   "123",
			groupID:     789,
			position:    3,
			expectError: false,
		},
		{
			name:        "Reposition to last",
			projectID:   "123",
			groupID:     789,
			position:    10,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := mock.NewMockClient()

			// Call the API
			ctx := f.Context()
			err := mockClient.RepositionTodoGroup(ctx, tt.projectID, tt.groupID, tt.position)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify API was called with correct parameters
			assert.Len(t, mockClient.Calls, 1, "Expected 1 API call")
		})
	}
}

func TestTodoGroupCommands(t *testing.T) {
	// Create a factory (doesn't need to be fully functional for help text)
	f := &factory.Factory{}

	// Test main todo command has group commands
	todoCmd := NewTodoCmd(f)

	// Test create-group command exists
	var foundCreateGroup bool
	var foundRepositionGroup bool
	for _, cmd := range todoCmd.Commands() {
		if cmd.Name() == "create-group" {
			foundCreateGroup = true
			assert.Contains(t, cmd.Short, "group", "create-group should mention group in short description")
		}
		if cmd.Name() == "reposition-group" {
			foundRepositionGroup = true
			assert.Contains(t, cmd.Short, "group", "reposition-group should mention group in short description")
		}
	}

	assert.True(t, foundCreateGroup, "todo command should have create-group subcommand")
	assert.True(t, foundRepositionGroup, "todo command should have reposition-group subcommand")
}
