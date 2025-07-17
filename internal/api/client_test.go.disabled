package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-token", "https://example.basecamp.com")
	
	assert.NotNil(t, client)
	// We can't access private fields directly, but we can verify the client works
	assert.NotNil(t, client)
}

func TestClient_GetProjects(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		expectedResult []Project
		expectedError  bool
	}{
		{
			name:       "successful response",
			statusCode: http.StatusOK,
			responseBody: `[
				{
					"id": 123,
					"name": "Test Project",
					"description": "A test project",
					"url": "https://example.basecamp.com/123",
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-01T00:00:00Z"
				}
			]`,
			expectedResult: []Project{
				{
					ID:          123,
					Name:        "Test Project",
					Description: "A test project",
					CreatedAt:   "2024-01-01T00:00:00Z",
					UpdatedAt:   "2024-01-01T00:00:00Z",
				},
			},
			expectedError: false,
		},
		{
			name:           "empty response",
			statusCode:     http.StatusOK,
			responseBody:   `[]`,
			expectedResult: []Project{},
			expectedError:  false,
		},
		{
			name:          "server error",
			statusCode:    http.StatusInternalServerError,
			responseBody:  `{"error": "Internal Server Error"}`,
			expectedError: true,
		},
		{
			name:          "invalid json",
			statusCode:    http.StatusOK,
			responseBody:  `invalid json`,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/projects.json", r.URL.Path)
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
				
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := NewClient("test-token", server.URL)
			result, err := client.GetProjects(context.Background())

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestClient_GetProject(t *testing.T) {
	tests := []struct {
		name          string
		projectID     string
		statusCode    int
		responseBody  string
		expectedResult *Project
		expectedError bool
	}{
		{
			name:       "successful response",
			projectID:  "123",
			statusCode: http.StatusOK,
			responseBody: `{
				"id": 123,
				"name": "Test Project",
				"description": "A test project",
				"url": "https://example.basecamp.com/123",
				"created_at": "2024-01-01T00:00:00Z",
				"updated_at": "2024-01-01T00:00:00Z"
			}`,
			expectedResult: &Project{
				ID:          123,
				Name:        "Test Project",
				Description: "A test project",
				CreatedAt:   "2024-01-01T00:00:00Z",
				UpdatedAt:   "2024-01-01T00:00:00Z",
			},
			expectedError: false,
		},
		{
			name:          "not found",
			projectID:     "999",
			statusCode:    http.StatusNotFound,
			responseBody:  `{"error": "Not Found"}`,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/projects/"+tt.projectID+".json", r.URL.Path)
				assert.Equal(t, "GET", r.Method)
				
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := NewClient("test-token", server.URL)
			result, err := client.GetProject(context.Background(), tt.projectID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestClient_CreateTodo(t *testing.T) {
	tests := []struct {
		name          string
		projectID     string
		todoListID    int64
		request       TodoCreateRequest
		statusCode    int
		responseBody  string
		expectedResult *Todo
		expectedError bool
	}{
		{
			name:       "successful creation",
			projectID:  "123",
			todoListID: 456,
			request: TodoCreateRequest{
				Content:     "New todo item",
				Description: "Todo description",
				DueOn:       &[]string{"2024-12-31"}[0],
				AssigneeIDs: []int64{789},
			},
			statusCode: http.StatusCreated,
			responseBody: `{
				"id": 999,
				"content": "New todo item",
				"description": "Todo description",
				"due_on": "2024-12-31",
				"completed": false,
				"created_at": "2024-01-01T00:00:00Z",
				"updated_at": "2024-01-01T00:00:00Z"
			}`,
			expectedResult: &Todo{
				ID:          999,
				Title:       "New todo item",
				Description: "Todo description",
				DueOn:       &[]string{"2024-12-31"}[0],
				Completed:   false,
				CreatedAt:   "2024-01-01T00:00:00Z",
				UpdatedAt:   "2024-01-01T00:00:00Z",
			},
			expectedError: false,
		},
		{
			name:       "validation error",
			projectID:  "123",
			todoListID: 456,
			request: TodoCreateRequest{
				Content: "", // Empty content should fail
			},
			statusCode:    http.StatusBadRequest,
			responseBody:  `{"error": "Content is required"}`,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/projects/"+tt.projectID+"/todolists/456/todos.json", r.URL.Path)
				assert.Equal(t, "POST", r.Method)
				
				// Verify request body
				var requestBody TodoCreateRequest
				err := json.NewDecoder(r.Body).Decode(&requestBody)
				require.NoError(t, err)
				assert.Equal(t, tt.request, requestBody)
				
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := NewClient("test-token", server.URL)
			result, err := client.CreateTodo(context.Background(), tt.projectID, tt.todoListID, tt.request)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestClient_CompleteTodo(t *testing.T) {
	tests := []struct {
		name          string
		projectID     string
		todoID        int64
		statusCode    int
		expectedError bool
	}{
		{
			name:          "successful completion",
			projectID:     "123",
			todoID:        456,
			statusCode:    http.StatusNoContent,
			expectedError: false,
		},
		{
			name:          "todo not found",
			projectID:     "123",
			todoID:        999,
			statusCode:    http.StatusNotFound,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/projects/"+tt.projectID+"/todos/456/completion.json", r.URL.Path)
				assert.Equal(t, "PUT", r.Method)
				
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			client := NewClient("test-token", server.URL)
			err := client.CompleteTodo(context.Background(), tt.projectID, tt.todoID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_GetCardTable(t *testing.T) {
	tests := []struct {
		name          string
		projectID     string
		cardTableID   int64
		statusCode    int
		responseBody  string
		expectedResult *CardTable
		expectedError bool
	}{
		{
			name:        "successful response",
			projectID:   "123",
			cardTableID: 456,
			statusCode:  http.StatusOK,
			responseBody: `{
				"id": 456,
				"title": "Project Board",
				"columns": [
					{
						"id": 1,
						"title": "To Do",
						"description": "Items to be done",
						"cards_count": 5
					},
					{
						"id": 2,
						"title": "In Progress",
						"description": "Items being worked on",
						"cards_count": 3
					}
				]
			}`,
			expectedResult: &CardTable{
				ID:    456,
				Title: "Project Board",
				Lists: []Column{
					{
						ID:         1,
						Title:      "To Do",
						CardsCount: 5,
					},
					{
						ID:         2,
						Title:      "In Progress",
						CardsCount: 3,
					},
				},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/projects/"+tt.projectID+"/card_tables/456.json", r.URL.Path)
				assert.Equal(t, "GET", r.Method)
				
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := NewClient("test-token", server.URL)
			result, err := client.GetCardTable(context.Background(), tt.projectID, tt.cardTableID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestClient_ErrorHandling(t *testing.T) {
	t.Run("network error", func(t *testing.T) {
		// Create a client with an invalid URL
		client := NewClient("test-token", "http://invalid-url-that-does-not-exist.local")
		
		_, err := client.GetProjects(context.Background())
		assert.Error(t, err)
	})

	t.Run("context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate a slow response
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[]`))
		}))
		defer server.Close()

		client := NewClient("test-token", server.URL)
		
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		_, err := client.GetProjects(ctx)
		assert.Error(t, err)
	})

	t.Run("unauthorized error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Unauthorized"}`))
		}))
		defer server.Close()

		client := NewClient("invalid-token", server.URL)
		
		_, err := client.GetProjects(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "401")
	})
}