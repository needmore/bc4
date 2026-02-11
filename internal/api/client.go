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
	return NewClientWithRetryConfig(accountID, accessToken, DefaultRetryConfig())
}

// NewClientWithRetryConfig creates a new API client with custom retry configuration
func NewClientWithRetryConfig(accountID, accessToken string, retryConfig RetryConfig) *Client {
	transport := NewRetryableTransport(http.DefaultTransport, retryConfig)
	return &Client{
		accountID:   accountID,
		accessToken: accessToken,
		baseURL:     defaultBaseURL,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}
}

func (c *Client) getBaseURL() string {
	return fmt.Sprintf("%s/%s", c.baseURL, c.accountID)
}

func (c *Client) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	return c.doRequestContext(context.Background(), method, path, body)
}

func (c *Client) doRequestContext(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	return c.doRequestWithHeadersContext(ctx, method, path, body, map[string]string{
		"Content-Type": "application/json; charset=utf-8",
	})
}

func (c *Client) doRequestWithHeaders(method, path string, body io.Reader, headers map[string]string) (*http.Response, error) {
	return c.doRequestWithHeadersContext(context.Background(), method, path, body, headers)
}

func (c *Client) doRequestWithHeadersContext(ctx context.Context, method, path string, body io.Reader, headers map[string]string) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.getBaseURL(), path)

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
	req.Header.Set("User-Agent", version.UserAgent())
	for k, v := range headers {
		req.Header.Set(k, v)
	}

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

// ProjectCreateRequest represents the payload for creating a new project
type ProjectCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// ProjectUpdateRequest represents the payload for updating a project
type ProjectUpdateRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// CreateProject creates a new project
func (c *Client) CreateProject(ctx context.Context, req ProjectCreateRequest) (*Project, error) {
	var project Project

	path := "/projects.json"
	if err := c.Post(path, req, &project); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return &project, nil
}

// UpdateProject updates an existing project
func (c *Client) UpdateProject(ctx context.Context, projectID string, req ProjectUpdateRequest) (*Project, error) {
	var project Project

	path := fmt.Sprintf("/projects/%s.json", projectID)
	if err := c.Put(path, req, &project); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return &project, nil
}

// DeleteProject trashes a project
func (c *Client) DeleteProject(ctx context.Context, projectID string) error {
	path := fmt.Sprintf("/projects/%s.json", projectID)
	if err := c.Delete(path); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}

// ArchiveProject archives a project
func (c *Client) ArchiveProject(ctx context.Context, projectID string) error {
	path := fmt.Sprintf("/projects/%s/status/archived.json", projectID)
	if err := c.Put(path, nil, nil); err != nil {
		return fmt.Errorf("failed to archive project: %w", err)
	}

	return nil
}

// UnarchiveProject restores an archived project to active status
func (c *Client) UnarchiveProject(ctx context.Context, projectID string) error {
	path := fmt.Sprintf("/projects/%s/status/active.json", projectID)
	if err := c.Put(path, nil, nil); err != nil {
		return fmt.Errorf("failed to unarchive project: %w", err)
	}

	return nil
}

// CopyProject creates a new project from a template.
// Note: sourceProjectID must be the ID of a template project.
// Regular projects cannot be used as a source - they must be marked as templates in Basecamp.
func (c *Client) CopyProject(ctx context.Context, sourceProjectID string, name string, description string) (*Project, error) {
	var project Project

	path := fmt.Sprintf("/templates/%s/project_constructions.json", sourceProjectID)
	req := ProjectCreateRequest{
		Name:        name,
		Description: description,
	}
	if err := c.Post(path, req, &project); err != nil {
		return nil, fmt.Errorf("failed to copy project: %w", err)
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

// Company represents a Basecamp company/organization
type Company struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Person represents a Basecamp user
type Person struct {
	ID           int64    `json:"id"`
	Name         string   `json:"name"`
	EmailAddress string   `json:"email_address"`
	Title        string   `json:"title"`
	AvatarURL    string   `json:"avatar_url"`
	Company      *Company `json:"company,omitempty"`
	CreatedAt    string   `json:"created_at,omitempty"`
	Admin        bool     `json:"admin,omitempty"`
	Owner        bool     `json:"owner,omitempty"`
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

// TodoUpdateRequest represents the payload for updating an existing todo
type TodoUpdateRequest struct {
	Content                 string  `json:"content,omitempty"`
	Description             string  `json:"description,omitempty"`
	DueOn                   *string `json:"due_on,omitempty"`
	StartsOn                *string `json:"starts_on,omitempty"`
	AssigneeIDs             []int64 `json:"assignee_ids,omitempty"`
	CompletionSubscriberIDs []int64 `json:"completion_subscriber_ids,omitempty"`
}

// UpdateTodo updates an existing todo
func (c *Client) UpdateTodo(ctx context.Context, projectID string, todoID int64, req TodoUpdateRequest) (*Todo, error) {
	var todo Todo

	path := fmt.Sprintf("/buckets/%s/todos/%d.json", projectID, todoID)
	if err := c.Put(path, req, &todo); err != nil {
		return nil, fmt.Errorf("failed to update todo: %w", err)
	}

	return &todo, nil
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

// TodoListUpdateRequest represents the payload for updating an existing todo list
type TodoListUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateTodoList updates an existing todo list
func (c *Client) UpdateTodoList(ctx context.Context, projectID string, todoListID int64, req TodoListUpdateRequest) (*TodoList, error) {
	var todoList TodoList

	path := fmt.Sprintf("/buckets/%s/todolists/%d.json", projectID, todoListID)
	if err := c.Put(path, req, &todoList); err != nil {
		return nil, fmt.Errorf("failed to update todo list: %w", err)
	}

	return &todoList, nil
}

// TodoGroupCreateRequest represents the payload for creating a new todo group
type TodoGroupCreateRequest struct {
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}

// CreateTodoGroup creates a new group within a todo list
func (c *Client) CreateTodoGroup(ctx context.Context, projectID string, todoListID int64, req TodoGroupCreateRequest) (*TodoGroup, error) {
	var group TodoGroup

	path := fmt.Sprintf("/buckets/%s/todolists/%d/groups.json", projectID, todoListID)
	if err := c.Post(path, req, &group); err != nil {
		return nil, fmt.Errorf("failed to create todo group: %w", err)
	}

	return &group, nil
}

// TodoGroupRepositionRequest represents the payload for repositioning a group
type TodoGroupRepositionRequest struct {
	Position int `json:"position"`
}

// RepositionTodoGroup repositions a group within a todo list
func (c *Client) RepositionTodoGroup(ctx context.Context, projectID string, groupID int64, position int) error {
	req := TodoGroupRepositionRequest{
		Position: position,
	}

	path := fmt.Sprintf("/buckets/%s/todolists/groups/%d/position.json", projectID, groupID)
	if err := c.Put(path, req, nil); err != nil {
		return fmt.Errorf("failed to reposition todo group: %w", err)
	}

	return nil
}

// TodoPositionRequest represents the payload for repositioning a todo
type TodoPositionRequest struct {
	Position int `json:"position"`
}

// RepositionTodo repositions a todo within its list
func (c *Client) RepositionTodo(ctx context.Context, projectID string, todoID int64, position int) error {
	req := TodoPositionRequest{
		Position: position,
	}

	path := fmt.Sprintf("/buckets/%s/todos/%d/position.json", projectID, todoID)
	if err := c.Put(path, req, nil); err != nil {
		return fmt.Errorf("failed to reposition todo: %w", err)
	}

	return nil
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

// GetMyProfile fetches the current user's profile
func (c *Client) GetMyProfile(ctx context.Context) (*Person, error) {
	var person Person

	path := "/my/profile.json"
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch profile: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err := json.NewDecoder(resp.Body).Decode(&person); err != nil {
		return nil, fmt.Errorf("failed to decode profile: %w", err)
	}

	return &person, nil
}

// GetAllPeople fetches all people visible to the current user in the account
func (c *Client) GetAllPeople(ctx context.Context) ([]Person, error) {
	var people []Person
	path := "/people.json"

	// Use paginated request to get all people
	pr := NewPaginatedRequest(c)
	if err := pr.GetAll(path, &people); err != nil {
		return nil, fmt.Errorf("failed to fetch people: %w", err)
	}

	return people, nil
}

// GetPingablePeople fetches all people who can be pinged in the account
func (c *Client) GetPingablePeople(ctx context.Context) ([]Person, error) {
	var people []Person
	path := "/circles/people.json"

	// Note: This endpoint is not paginated according to Basecamp API docs
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pingable people: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err := json.NewDecoder(resp.Body).Decode(&people); err != nil {
		return nil, fmt.Errorf("failed to decode pingable people: %w", err)
	}

	return people, nil
}

// ProjectAccessUpdateRequest represents the payload for updating project access
type ProjectAccessUpdateRequest struct {
	Grant  []int64                  `json:"grant,omitempty"`
	Revoke []int64                  `json:"revoke,omitempty"`
	Create []ProjectAccessNewPerson `json:"create,omitempty"`
}

// ProjectAccessNewPerson represents a new person to invite to a project
type ProjectAccessNewPerson struct {
	Name         string `json:"name"`
	EmailAddress string `json:"email_address"`
	Title        string `json:"title,omitempty"`
	CompanyName  string `json:"company_name,omitempty"`
}

// ProjectAccessUpdateResponse represents the response from updating project access
type ProjectAccessUpdateResponse struct {
	Granted []Person `json:"granted"`
	Revoked []Person `json:"revoked"`
}

// UpdateProjectAccess updates who has access to a project (grant, revoke, or create new users)
func (c *Client) UpdateProjectAccess(ctx context.Context, projectID string, req ProjectAccessUpdateRequest) (*ProjectAccessUpdateResponse, error) {
	var response ProjectAccessUpdateResponse

	path := fmt.Sprintf("/projects/%s/people/users.json", projectID)
	if err := c.Put(path, req, &response); err != nil {
		return nil, fmt.Errorf("failed to update project access: %w", err)
	}

	return &response, nil
}
