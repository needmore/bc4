package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	_ "time"

	"github.com/stretchr/testify/assert"
)

func TestGetMessageBoard(t *testing.T) {
	tests := []struct {
		name          string
		projectID     string
		responseCode  int
		responseBody  string
		expectedError bool
		errorContains string
	}{
		{
			name:         "successful response",
			projectID:    "123456",
			responseCode: http.StatusOK,
			responseBody: `{
				"dock": [
					{
						"id": 789,
						"title": "Message Board",
						"name": "message_board",
						"url": "https://3.basecamp.com/123456/buckets/123456/message_boards/789"
					}
				]
			}`,
			expectedError: false,
		},
		{
			name:          "project not found",
			projectID:     "999999",
			responseCode:  http.StatusNotFound,
			responseBody:  `{"error": "Project not found"}`,
			expectedError: true,
			errorContains: "not found",
		},
		{
			name:         "message board not in dock",
			projectID:    "123456",
			responseCode: http.StatusOK,
			responseBody: `{
				"dock": [
					{
						"id": 456,
						"title": "To-dos",
						"name": "todoset",
						"url": "https://3.basecamp.com/123456/buckets/123456/todosets/456"
					}
				]
			}`,
			expectedError: true,
			errorContains: "message board not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Track request count for this test
			requestCount := 0
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCount++
				
				// Handle different requests based on path and count
				if r.URL.Path == "/123456/projects/123456.json" {
					if requestCount == 1 {
						// First request: return basic project
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"id": 123456, "name": "Test Project"}`))
					} else if requestCount == 2 && tt.name == "successful response" {
						// Second request for successful test: return project with dock
						w.WriteHeader(tt.responseCode)
						w.Write([]byte(tt.responseBody))
					} else {
						// Other cases
						w.WriteHeader(tt.responseCode)
						w.Write([]byte(tt.responseBody))
					}
					return
				}
				
				// Message board details request
				if r.URL.Path == "/123456/buckets/123456/message_boards/789.json" {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{
						"id": 789,
						"title": "Message Board",
						"name": "message_board",
						"status": "active",
						"created_at": "2023-01-01T00:00:00Z",
						"updated_at": "2023-01-01T00:00:00Z"
					}`))
					return
				}
				
				// Default response
				w.WriteHeader(tt.responseCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Create client
			client := &Client{
				accountID:  "123456",
				baseURL:    server.URL,
				httpClient: &http.Client{},
			}

			// Test
			board, err := client.GetMessageBoard(context.Background(), tt.projectID)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, board)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, board)
				if board != nil {
					assert.Equal(t, int64(789), board.ID)
					assert.Equal(t, "Message Board", board.Title)
				}
			}
		})
	}
}

func TestListMessages(t *testing.T) {
	tests := []struct {
		name           string
		projectID      string
		messageBoardID int64
		responseCode   int
		responseBody   string
		expectedCount  int
		expectedError  bool
	}{
		{
			name:           "successful response",
			projectID:      "123456",
			messageBoardID: 789,
			responseCode:   http.StatusOK,
			responseBody: `[
				{
					"id": 1,
					"subject": "First Message",
					"content": "<div>Hello World</div>",
					"status": "active",
					"created_at": "2023-01-01T00:00:00Z",
					"updated_at": "2023-01-01T00:00:00Z",
					"creator": {"id": 100, "name": "John Doe"},
					"comments_count": 5
				},
				{
					"id": 2,
					"subject": "Second Message",
					"content": "<div>Another message</div>",
					"status": "active",
					"created_at": "2023-01-02T00:00:00Z",
					"updated_at": "2023-01-02T00:00:00Z",
					"creator": {"id": 101, "name": "Jane Smith"},
					"comments_count": 2
				}
			]`,
			expectedCount: 2,
			expectedError: false,
		},
		{
			name:           "empty response",
			projectID:      "123456",
			messageBoardID: 789,
			responseCode:   http.StatusOK,
			responseBody:   `[]`,
			expectedCount:  0,
			expectedError:  false,
		},
		{
			name:           "error response",
			projectID:      "123456",
			messageBoardID: 789,
			responseCode:   http.StatusNotFound,
			responseBody:   `{"error": "Not found"}`,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/123456/buckets/123456/message_boards/789/messages.json", r.URL.Path)
				w.WriteHeader(tt.responseCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Create client
			client := &Client{
				accountID:  "123456",
				baseURL:    server.URL,
				httpClient: &http.Client{},
			}

			// Test
			messages, err := client.ListMessages(context.Background(), tt.projectID, tt.messageBoardID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, messages)
			} else {
				assert.NoError(t, err)
				// messages might be nil or empty slice depending on PaginatedRequest behavior
				if messages != nil {
					assert.Len(t, messages, tt.expectedCount)
				}

				if tt.expectedCount > 0 {
					assert.Equal(t, "First Message", messages[0].Subject)
					assert.Equal(t, "John Doe", messages[0].Creator.Name)
					assert.Equal(t, 5, messages[0].CommentsCount)
				}
			}
		})
	}
}

func TestCreateMessage(t *testing.T) {
	tests := []struct {
		name          string
		request       MessageCreateRequest
		responseCode  int
		responseBody  string
		expectedID    int64
		expectedError bool
	}{
		{
			name: "successful creation",
			request: MessageCreateRequest{
				Subject: "New Message",
				Content: "<div>Message content</div>",
				Status:  "active",
			},
			responseCode: http.StatusCreated,
			responseBody: `{
				"id": 123,
				"subject": "New Message",
				"content": "<div>Message content</div>",
				"status": "active",
				"created_at": "2023-01-01T00:00:00Z",
				"updated_at": "2023-01-01T00:00:00Z",
				"creator": {"id": 100, "name": "John Doe"}
			}`,
			expectedID:    123,
			expectedError: false,
		},
		{
			name: "draft message",
			request: MessageCreateRequest{
				Subject: "Draft Message",
				Content: "<div>Draft content</div>",
				Status:  "draft",
			},
			responseCode: http.StatusCreated,
			responseBody: `{
				"id": 124,
				"subject": "Draft Message",
				"content": "<div>Draft content</div>",
				"status": "draft",
				"created_at": "2023-01-01T00:00:00Z",
				"updated_at": "2023-01-01T00:00:00Z",
				"creator": {"id": 100, "name": "John Doe"}
			}`,
			expectedID:    124,
			expectedError: false,
		},
		{
			name: "validation error",
			request: MessageCreateRequest{
				Subject: "",
				Content: "",
			},
			responseCode:  http.StatusBadRequest,
			responseBody:  `{"error": "Subject is required"}`,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/123456/buckets/123456/message_boards/789/messages.json", r.URL.Path)

				// Verify request body
				var req MessageCreateRequest
				err := json.NewDecoder(r.Body).Decode(&req)
				assert.NoError(t, err)
				assert.Equal(t, tt.request.Subject, req.Subject)
				assert.Equal(t, tt.request.Content, req.Content)

				w.WriteHeader(tt.responseCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Create client
			client := &Client{
				accountID:  "123456",
				baseURL:    server.URL,
				httpClient: &http.Client{},
			}

			// Test
			message, err := client.CreateMessage(context.Background(), "123456", 789, tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, message)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, message)
				assert.Equal(t, tt.expectedID, message.ID)
				assert.Equal(t, tt.request.Subject, message.Subject)
			}
		})
	}
}

func TestUpdateMessage(t *testing.T) {
	tests := []struct {
		name          string
		messageID     int64
		request       MessageUpdateRequest
		responseCode  int
		expectedError bool
	}{
		{
			name:      "successful update",
			messageID: 123,
			request: MessageUpdateRequest{
				Subject: "Updated Subject",
				Content: "<div>Updated content</div>",
			},
			responseCode:  http.StatusOK,
			expectedError: false,
		},
		{
			name:      "not found",
			messageID: 999,
			request: MessageUpdateRequest{
				Subject: "Updated Subject",
			},
			responseCode:  http.StatusNotFound,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)
				w.WriteHeader(tt.responseCode)
				if !tt.expectedError {
					w.Write([]byte(`{
						"id": 123,
						"subject": "Updated Subject",
						"content": "<div>Updated content</div>"
					}`))
				}
			}))
			defer server.Close()

			// Create client
			client := &Client{
				accountID:  "123456",
				baseURL:    server.URL,
				httpClient: &http.Client{},
			}

			// Test
			message, err := client.UpdateMessage(context.Background(), "123456", tt.messageID, tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, message)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, message)
			}
		})
	}
}

