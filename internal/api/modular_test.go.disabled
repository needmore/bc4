package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModularClient_Projects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/projects.json":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"id": 123, "name": "Test Project"}]`))
		case "/projects/123.json":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id": 123, "name": "Test Project"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewModularClient("test-token", server.URL)
	projectOps := client.Projects()

	t.Run("get projects", func(t *testing.T) {
		projects, err := projectOps.GetProjects(context.Background())
		require.NoError(t, err)
		assert.Len(t, projects, 1)
		assert.Equal(t, int64(123), projects[0].ID)
	})

	t.Run("get project", func(t *testing.T) {
		project, err := projectOps.GetProject(context.Background(), "123")
		require.NoError(t, err)
		assert.Equal(t, int64(123), project.ID)
		assert.Equal(t, "Test Project", project.Name)
	})

	t.Run("operations return same instance", func(t *testing.T) {
		// Verify that calling Projects() multiple times returns the same instance
		ops1 := client.Projects()
		ops2 := client.Projects()
		assert.Equal(t, ops1, ops2)
	})
}

func TestModularClient_Todos(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/projects/123/todosets/456.json":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id": 456, "name": "Todo Set"}`))
		case r.URL.Path == "/projects/123/todolists/789/todos.json" && r.Method == "GET":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"id": 999, "title": "Test Todo"}]`))
		case r.URL.Path == "/projects/123/todolists/789/todos.json" && r.Method == "POST":
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id": 1000, "title": "New Todo"}`))
		case r.URL.Path == "/projects/123/todos/999/completion.json":
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewModularClient("test-token", server.URL)
	todoOps := client.Todos()

	t.Run("get project todo set", func(t *testing.T) {
		todoSet, err := todoOps.GetProjectTodoSet(context.Background(), "123")
		require.NoError(t, err)
		assert.Equal(t, int64(456), todoSet.ID)
	})

	t.Run("get todos", func(t *testing.T) {
		todos, err := todoOps.GetTodos(context.Background(), "123", 789)
		require.NoError(t, err)
		assert.Len(t, todos, 1)
		assert.Equal(t, "Test Todo", todos[0].Title)
	})

	t.Run("create todo", func(t *testing.T) {
		req := TodoCreateRequest{Content: "New Todo"}
		todo, err := todoOps.CreateTodo(context.Background(), "123", 789, req)
		require.NoError(t, err)
		assert.Equal(t, int64(1000), todo.ID)
		assert.Equal(t, "New Todo", todo.Title)
	})

	t.Run("complete todo", func(t *testing.T) {
		err := todoOps.CompleteTodo(context.Background(), "123", 999)
		assert.NoError(t, err)
	})
}

func TestModularClient_Cards(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/projects/123/card_tables/456.json":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": 456,
				"title": "Board",
				"columns": [{"id": 1, "title": "To Do"}]
			}`))
		case r.URL.Path == "/projects/123/columns/1/cards.json":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"id": 789, "title": "Test Card"}]`))
		case r.URL.Path == "/projects/123/cards/789.json" && r.Method == "GET":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id": 789, "title": "Test Card", "content": "Card content"}`))
		case r.URL.Path == "/projects/123/columns/1/cards.json" && r.Method == "POST":
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id": 790, "title": "New Card"}`))
		case r.URL.Path == "/projects/123/cards/789/moves.json":
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewModularClient("test-token", server.URL)
	cardOps := client.Cards()

	t.Run("get card table", func(t *testing.T) {
		table, err := cardOps.GetCardTable(context.Background(), "123", 456)
		require.NoError(t, err)
		assert.Equal(t, "Board", table.Title)
		assert.Len(t, table.Lists, 1)
	})

	t.Run("get cards in column", func(t *testing.T) {
		cards, err := cardOps.GetCardsInColumn(context.Background(), "123", 1)
		require.NoError(t, err)
		assert.Len(t, cards, 1)
		assert.Equal(t, "Test Card", cards[0].Title)
	})

	t.Run("get card", func(t *testing.T) {
		card, err := cardOps.GetCard(context.Background(), "123", 789)
		require.NoError(t, err)
		assert.Equal(t, "Test Card", card.Title)
		assert.Equal(t, "Card content", card.Content)
	})

	t.Run("create card", func(t *testing.T) {
		req := CardCreateRequest{Title: "New Card"}
		card, err := cardOps.CreateCard(context.Background(), "123", 1, req)
		require.NoError(t, err)
		assert.Equal(t, int64(790), card.ID)
	})

	t.Run("move card", func(t *testing.T) {
		err := cardOps.MoveCard(context.Background(), "123", 789, 2)
		assert.NoError(t, err)
	})
}

func TestModularClient_Steps(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/projects/123/cards/456/steps.json" && r.Method == "POST":
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id": 789, "title": "New Step"}`))
		case r.URL.Path == "/projects/123/steps/789.json" && r.Method == "PUT":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id": 789, "title": "Updated Step"}`))
		case r.URL.Path == "/projects/123/steps/789/completion.json":
			w.WriteHeader(http.StatusNoContent)
		case r.URL.Path == "/projects/123/steps/789.json" && r.Method == "DELETE":
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewModularClient("test-token", server.URL)
	stepOps := client.Steps()

	t.Run("create step", func(t *testing.T) {
		req := StepCreateRequest{Title: "New Step"}
		step, err := stepOps.CreateStep(context.Background(), "123", 456, req)
		require.NoError(t, err)
		assert.Equal(t, int64(789), step.ID)
	})

	t.Run("update step", func(t *testing.T) {
		req := StepUpdateRequest{Title: "Updated Step"}
		step, err := stepOps.UpdateStep(context.Background(), "123", 789, req)
		require.NoError(t, err)
		assert.Equal(t, "Updated Step", step.Title)
	})

	t.Run("set step completion", func(t *testing.T) {
		err := stepOps.SetStepCompletion(context.Background(), "123", 789, true)
		assert.NoError(t, err)
	})

	t.Run("delete step", func(t *testing.T) {
		err := stepOps.DeleteStep(context.Background(), "123", 789)
		assert.NoError(t, err)
	})
}

func TestModularClient_AllOperations(t *testing.T) {
	// Test that all operation methods return non-nil interfaces
	client := NewModularClient("test-token", "https://example.com")

	assert.NotNil(t, client.Projects())
	assert.NotNil(t, client.Todos())
	assert.NotNil(t, client.Campfires())
	assert.NotNil(t, client.Cards())
	assert.NotNil(t, client.Steps())
	assert.NotNil(t, client.Columns())
	assert.NotNil(t, client.People())
}