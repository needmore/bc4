package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/needmore/bc4/internal/errors"
	"github.com/needmore/bc4/internal/version"
)

const (
	defaultBaseURL = "https://3.basecampapi.com"
)

type Client struct {
	accountID   string
	accessToken string
	httpClient  *http.Client
	baseURL     string
}

// NewClient creates a new API client
// Deprecated: Use NewModularClient instead for better separation of concerns
func NewClient(accountID, accessToken string) *Client {
	return &Client{
		accountID:   accountID,
		accessToken: accessToken,
		baseURL:     defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) getBaseURL() string {
	return fmt.Sprintf("%s/%s", c.baseURL, c.accountID)
}

func (c *Client) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.getBaseURL(), path)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", version.UserAgent())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer func() { _ = resp.Body.Close() }()
		body, _ := io.ReadAll(resp.Body)

		// Use our custom error types for better user experience
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return nil, errors.NewAuthenticationError(fmt.Errorf("unauthorized: %s", string(body)))
		case http.StatusNotFound:
			// Try to extract resource info from the path
			parts := strings.Split(path, "/")
			resource := "resource"
			if len(parts) > 2 {
				// Extract resource type from path (e.g., /buckets/123/projects/456 -> project)
				resource = strings.TrimSuffix(parts[len(parts)-2], "s")
			}
			return nil, errors.NewNotFoundError(resource, "", fmt.Errorf("not found: %s", string(body)))
		default:
			return nil, errors.NewAPIError(resp.StatusCode, string(body), nil)
		}
	}

	return resp, nil
}

func (c *Client) Get(path string, result interface{}) error {
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

func (c *Client) Post(path string, payload interface{}, result interface{}) error {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = strings.NewReader(string(jsonData))
	}

	resp, err := c.doRequest("POST", path, body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

func (c *Client) Put(path string, payload interface{}, result interface{}) error {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = strings.NewReader(string(jsonData))
	}

	resp, err := c.doRequest("PUT", path, body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

func (c *Client) Delete(path string) error {
	resp, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	return nil
}

// Project represents a Basecamp project
type Project struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Purpose     string `json:"purpose"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// GetProjects fetches all projects for the account (handles pagination)
func (c *Client) GetProjects(ctx context.Context) ([]Project, error) {
	var projects []Project

	// Use paginated request to get all projects
	pr := NewPaginatedRequest(c)
	if err := pr.GetAll("/projects.json", &projects); err != nil {
		return nil, fmt.Errorf("failed to fetch projects: %w", err)
	}

	return projects, nil
}

// GetProject fetches a single project by ID
func (c *Client) GetProject(ctx context.Context, projectID string) (*Project, error) {
	var project Project

	path := fmt.Sprintf("/projects/%s.json", projectID)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch project: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, fmt.Errorf("failed to decode project: %w", err)
	}

	return &project, nil
}

// TodoSet represents a Basecamp todo set (container for todo lists)
type TodoSet struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	Name         string `json:"name"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	TodolistsURL string `json:"todolists_url"`
}

// TodoList represents a Basecamp todo list
type TodoList struct {
	ID             int64  `json:"id"`
	Title          string `json:"title"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
	Completed      bool   `json:"completed"`
	CompletedRatio string `json:"completed_ratio"`
	TodosCount     int    `json:"todos_count"`
	TodosURL       string `json:"todos_url"`
	GroupsURL      string `json:"groups_url"`
}

// TodoGroup represents a group of todos within a todo list
type TodoGroup struct {
	ID             int64  `json:"id"`
	Title          string `json:"title"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
	Completed      bool   `json:"completed"`
	CompletedRatio string `json:"completed_ratio"`
	TodosCount     int    `json:"todos_count"`
	TodosURL       string `json:"todos_url"`
	Position       int    `json:"position"`
}

// Person represents a Basecamp user
type Person struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	EmailAddress string `json:"email_address"`
	Title        string `json:"title"`
	AvatarURL    string `json:"avatar_url"`
}

// Todo represents a Basecamp todo item
type Todo struct {
	ID          int64    `json:"id"`
	Title       string   `json:"title"`
	Content     string   `json:"content"`
	Description string   `json:"description"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
	Completed   bool     `json:"completed"`
	DueOn       *string  `json:"due_on"`
	StartsOn    *string  `json:"starts_on"`
	TodolistID  int64    `json:"todolist_id"`
	Creator     *Person  `json:"creator"`
	Assignees   []Person `json:"assignees"`
}

// GetProjectTodoSet fetches the todo set for a project
func (c *Client) GetProjectTodoSet(ctx context.Context, projectID string) (*TodoSet, error) {
	// First get the project to find its todo set
	project, err := c.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Get project tools/features
	path := fmt.Sprintf("/projects/%d.json", project.ID)

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch project tools: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var projectData struct {
		Dock []struct {
			ID    int64  `json:"id"`
			Title string `json:"title"`
			Name  string `json:"name"`
			URL   string `json:"url"`
		} `json:"dock"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&projectData); err != nil {
		return nil, fmt.Errorf("failed to decode project data: %w", err)
	}

	// Find the todoset in the dock
	for _, tool := range projectData.Dock {
		if tool.Name == "todoset" {
			return &TodoSet{
				ID:           tool.ID,
				Title:        tool.Title,
				Name:         tool.Name,
				TodolistsURL: tool.URL,
			}, nil
		}
	}

	return nil, fmt.Errorf("todo set not found for project")
}

// GetTodoLists fetches all todo lists in a todo set
func (c *Client) GetTodoLists(ctx context.Context, projectID string, todoSetID int64) ([]TodoList, error) {
	var todoLists []TodoList
	path := fmt.Sprintf("/buckets/%s/todosets/%d/todolists.json", projectID, todoSetID)

	// Use paginated request to get all todo lists
	pr := NewPaginatedRequest(c)
	if err := pr.GetAll(path, &todoLists); err != nil {
		return nil, fmt.Errorf("failed to fetch todo lists: %w", err)
	}

	return todoLists, nil
}

// GetTodoList fetches a single todo list by ID
func (c *Client) GetTodoList(ctx context.Context, projectID string, todoListID int64) (*TodoList, error) {
	var todoList TodoList

	path := fmt.Sprintf("/buckets/%s/todolists/%d.json", projectID, todoListID)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch todo list: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err := json.NewDecoder(resp.Body).Decode(&todoList); err != nil {
		return nil, fmt.Errorf("failed to decode todo list: %w", err)
	}

	return &todoList, nil
}

// GetTodos fetches all todos in a todo list
func (c *Client) GetTodos(ctx context.Context, projectID string, todoListID int64) ([]Todo, error) {
	var todos []Todo
	path := fmt.Sprintf("/buckets/%s/todolists/%d/todos.json", projectID, todoListID)

	// Use paginated request to get all todos
	pr := NewPaginatedRequest(c)
	if err := pr.GetAll(path, &todos); err != nil {
		return nil, fmt.Errorf("failed to fetch todos: %w", err)
	}

	return todos, nil
}

// GetAllTodos fetches all todos in a todo list including completed ones
func (c *Client) GetAllTodos(ctx context.Context, projectID string, todoListID int64) ([]Todo, error) {
	var allTodos []Todo

	// Get incomplete todos
	incompleteTodos, err := c.GetTodos(ctx, projectID, todoListID)
	if err != nil {
		return nil, err
	}
	allTodos = append(allTodos, incompleteTodos...)

	// Get completed todos using the completed=true parameter
	var completedTodos []Todo
	path := fmt.Sprintf("/buckets/%s/todolists/%d/todos.json?completed=true", projectID, todoListID)

	// Use paginated request to get all completed todos
	pr := NewPaginatedRequest(c)
	if err := pr.GetAll(path, &completedTodos); err != nil {
		// If we can't get completed todos, just return the incomplete ones
		return allTodos, err
	}

	// Mark them as completed (in case the API doesn't set this)
	for i := range completedTodos {
		completedTodos[i].Completed = true
	}

	allTodos = append(allTodos, completedTodos...)
	return allTodos, nil
}

// GetTodoGroups fetches all groups in a todo list
func (c *Client) GetTodoGroups(ctx context.Context, projectID string, todoListID int64) ([]TodoGroup, error) {
	var groups []TodoGroup
	path := fmt.Sprintf("/buckets/%s/todolists/%d/groups.json", projectID, todoListID)

	// Use paginated request to get all groups
	pr := NewPaginatedRequest(c)
	if err := pr.GetAll(path, &groups); err != nil {
		return nil, fmt.Errorf("failed to fetch todo groups: %w", err)
	}

	return groups, nil
}

// TodoCreateRequest represents the payload for creating a new todo
type TodoCreateRequest struct {
	Content     string  `json:"content"`
	Description string  `json:"description,omitempty"`
	DueOn       *string `json:"due_on,omitempty"`
	StartsOn    *string `json:"starts_on,omitempty"`
	AssigneeIDs []int64 `json:"assignee_ids,omitempty"`
}

// CreateTodo creates a new todo in a todo list
func (c *Client) CreateTodo(ctx context.Context, projectID string, todoListID int64, req TodoCreateRequest) (*Todo, error) {
	var todo Todo

	path := fmt.Sprintf("/buckets/%s/todolists/%d/todos.json", projectID, todoListID)
	if err := c.Post(path, req, &todo); err != nil {
		return nil, fmt.Errorf("failed to create todo: %w", err)
	}

	return &todo, nil
}

// CompleteTodo marks a todo as complete
func (c *Client) CompleteTodo(ctx context.Context, projectID string, todoID int64) error {
	path := fmt.Sprintf("/buckets/%s/todos/%d/completion.json", projectID, todoID)
	if err := c.Put(path, nil, nil); err != nil {
		return fmt.Errorf("failed to complete todo: %w", err)
	}

	return nil
}

// UncompleteTodo marks a todo as incomplete
func (c *Client) UncompleteTodo(ctx context.Context, projectID string, todoID int64) error {
	path := fmt.Sprintf("/buckets/%s/todos/%d/completion.json", projectID, todoID)
	if err := c.Delete(path); err != nil {
		return fmt.Errorf("failed to uncomplete todo: %w", err)
	}

	return nil
}

// TodoListCreateRequest represents the payload for creating a new todo list
type TodoListCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// CreateTodoList creates a new todo list in a project
func (c *Client) CreateTodoList(ctx context.Context, projectID string, todoSetID int64, req TodoListCreateRequest) (*TodoList, error) {
	var todoList TodoList

	path := fmt.Sprintf("/buckets/%s/todosets/%d/todolists.json", projectID, todoSetID)
	if err := c.Post(path, req, &todoList); err != nil {
		return nil, fmt.Errorf("failed to create todo list: %w", err)
	}

	return &todoList, nil
}

// GetTodo fetches a single todo by ID
func (c *Client) GetTodo(ctx context.Context, projectID string, todoID int64) (*Todo, error) {
	var todo Todo

	path := fmt.Sprintf("/buckets/%s/todos/%d.json", projectID, todoID)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch todo: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err := json.NewDecoder(resp.Body).Decode(&todo); err != nil {
		return nil, fmt.Errorf("failed to decode todo: %w", err)
	}

	return &todo, nil
}

// GetProjectPeople fetches all people associated with a project
func (c *Client) GetProjectPeople(ctx context.Context, projectID string) ([]Person, error) {
	var people []Person
	path := fmt.Sprintf("/projects/%s/people.json", projectID)

	// Use paginated request to get all people
	pr := NewPaginatedRequest(c)
	if err := pr.GetAll(path, &people); err != nil {
		return nil, fmt.Errorf("failed to fetch project people: %w", err)
	}

	return people, nil
}

// GetPerson fetches a specific person by ID
func (c *Client) GetPerson(ctx context.Context, personID int64) (*Person, error) {
	var person Person

	path := fmt.Sprintf("/people/%d.json", personID)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch person: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err := json.NewDecoder(resp.Body).Decode(&person); err != nil {
		return nil, fmt.Errorf("failed to decode person: %w", err)
	}

	return &person, nil
}
