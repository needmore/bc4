package mock

import (
	"context"
	"errors"
	"testing"

	"github.com/needmore/bc4/internal/api"
	"github.com/stretchr/testify/assert"
)

func TestNewMockClient(t *testing.T) {
	client := NewMockClient()
	assert.NotNil(t, client)
	assert.NotNil(t, client.Calls)
	assert.Empty(t, client.Calls)
}

func TestMockClient_GetProjects(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		client := NewMockClient()
		expectedProjects := []api.Project{
			{ID: 1, Name: "Project 1"},
			{ID: 2, Name: "Project 2"},
		}
		client.Projects = expectedProjects

		result, err := client.GetProjects(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, expectedProjects, result)
		assert.Contains(t, client.Calls, "GetProjects")
	})

	t.Run("error response", func(t *testing.T) {
		client := NewMockClient()
		client.ProjectsError = errors.New("test error")

		result, err := client.GetProjects(context.Background())

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "test error", err.Error())
		assert.Contains(t, client.Calls, "GetProjects")
	})
}

func TestMockClient_GetProject(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		client := NewMockClient()
		expectedProject := &api.Project{ID: 123, Name: "Test Project"}
		client.Project = expectedProject

		result, err := client.GetProject(context.Background(), "123")

		assert.NoError(t, err)
		assert.Equal(t, expectedProject, result)
		assert.Contains(t, client.Calls, "GetProject(123)")
	})

	t.Run("error response", func(t *testing.T) {
		client := NewMockClient()
		client.ProjectError = errors.New("custom error")

		result, err := client.GetProject(context.Background(), "123")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, client.Calls, "GetProject(123)")
	})

	t.Run("nil project returns error", func(t *testing.T) {
		client := NewMockClient()
		client.Project = nil

		result, err := client.GetProject(context.Background(), "123")

		assert.Error(t, err)
		assert.Equal(t, "project not found", err.Error())
		assert.Nil(t, result)
	})
}

func TestMockClient_TodoOperations(t *testing.T) {
	t.Run("create todo", func(t *testing.T) {
		client := NewMockClient()
		expectedTodo := &api.Todo{ID: 999, Content: "New Todo"}
		client.CreatedTodo = expectedTodo

		req := api.TodoCreateRequest{Content: "New Todo"}
		result, err := client.CreateTodo(context.Background(), "123", 456, req)

		assert.NoError(t, err)
		assert.Equal(t, expectedTodo, result)
		assert.Contains(t, client.Calls[0], "CreateTodo(123, 456")
	})

	t.Run("complete todo", func(t *testing.T) {
		client := NewMockClient()

		err := client.CompleteTodo(context.Background(), "123", 456)

		assert.NoError(t, err)
		assert.Contains(t, client.Calls, "CompleteTodo(123, 456)")
	})

	t.Run("complete todo with error", func(t *testing.T) {
		client := NewMockClient()
		client.CompleteTodoError = errors.New("completion failed")

		err := client.CompleteTodo(context.Background(), "123", 456)

		assert.Error(t, err)
		assert.Equal(t, "completion failed", err.Error())
	})
}

func TestMockClient_CardOperations(t *testing.T) {
	t.Run("get card table", func(t *testing.T) {
		client := NewMockClient()
		expectedTable := &api.CardTable{
			ID:    456,
			Title: "Test Board",
			Lists: []api.Column{
				{ID: 1, Title: "To Do"},
			},
		}
		client.CardTable = expectedTable

		result, err := client.GetCardTable(context.Background(), "123", 456)

		assert.NoError(t, err)
		assert.Equal(t, expectedTable, result)
		assert.Contains(t, client.Calls, "GetCardTable(123, 456)")
	})

	t.Run("create card", func(t *testing.T) {
		client := NewMockClient()
		expectedCard := &api.Card{ID: 789, Title: "New Card"}
		client.CreatedCard = expectedCard

		req := api.CardCreateRequest{Title: "New Card"}
		result, err := client.CreateCard(context.Background(), "123", 1, req)

		assert.NoError(t, err)
		assert.Equal(t, expectedCard, result)
		assert.Contains(t, client.Calls[0], "CreateCard(123, 1")
	})

	t.Run("move card", func(t *testing.T) {
		client := NewMockClient()

		err := client.MoveCard(context.Background(), "123", 789, 2)

		assert.NoError(t, err)
		assert.Contains(t, client.Calls, "MoveCard(123, 789, 2)")
	})
}

func TestMockClient_StepOperations(t *testing.T) {
	t.Run("create step", func(t *testing.T) {
		client := NewMockClient()
		expectedStep := &api.Step{ID: 999, Title: "New Step"}
		client.CreatedStep = expectedStep

		req := api.StepCreateRequest{Title: "New Step"}
		result, err := client.CreateStep(context.Background(), "123", 456, req)

		assert.NoError(t, err)
		assert.Equal(t, expectedStep, result)
		assert.Contains(t, client.Calls[0], "CreateStep(123, 456")
	})

	t.Run("set step completion", func(t *testing.T) {
		client := NewMockClient()

		err := client.SetStepCompletion(context.Background(), "123", 999, true)

		assert.NoError(t, err)
		assert.Contains(t, client.Calls, "SetStepCompletion(123, 999, true)")
	})

	t.Run("delete step", func(t *testing.T) {
		client := NewMockClient()

		err := client.DeleteStep(context.Background(), "123", 999)

		assert.NoError(t, err)
		assert.Contains(t, client.Calls, "DeleteStep(123, 999)")
	})
}

func TestMockClient_CallTracking(t *testing.T) {
	client := NewMockClient()

	// Make several calls
	_, _ = client.GetProjects(context.Background())
	_, _ = client.GetProject(context.Background(), "123")
	_, _ = client.GetTodo(context.Background(), "123", 456)
	_ = client.CompleteTodo(context.Background(), "123", 456)

	// Verify all calls were tracked
	assert.Len(t, client.Calls, 4)
	assert.Equal(t, "GetProjects", client.Calls[0])
	assert.Equal(t, "GetProject(123)", client.Calls[1])
	assert.Equal(t, "GetTodo(123, 456)", client.Calls[2])
	assert.Equal(t, "CompleteTodo(123, 456)", client.Calls[3])
}

func TestMockClient_ImplementsInterface(t *testing.T) {
	// This test ensures the mock implements the interface correctly
	// The compile-time check in the mock file ensures this, but this
	// test documents the requirement
	var _ api.APIClient = (*MockClient)(nil)
}
