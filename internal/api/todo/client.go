package todo

import (
	"context"
	"fmt"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/api/project"
)

// Client is the interface for todo-related API calls
type Client interface {
	GetProjectTodoSet(ctx context.Context, projectID string) (*api.TodoSet, error)
	GetTodoLists(ctx context.Context, projectID string, todoSetID int64) ([]api.TodoList, error)
	GetTodoList(ctx context.Context, projectID string, todoListID int64) (*api.TodoList, error)
	GetTodos(ctx context.Context, projectID string, todoListID int64) ([]api.Todo, error)
	GetAllTodos(ctx context.Context, projectID string, todoListID int64) ([]api.Todo, error)
	GetTodo(ctx context.Context, projectID string, todoID int64) (*api.Todo, error)
	GetTodoGroups(ctx context.Context, projectID string, todoListID int64) ([]api.TodoGroup, error)
	CreateTodo(ctx context.Context, projectID string, todoListID int64, req api.TodoCreateRequest) (*api.Todo, error)
	CreateTodoList(ctx context.Context, projectID string, todoSetID int64, req api.TodoListCreateRequest) (*api.TodoList, error)
	CompleteTodo(ctx context.Context, projectID string, todoID int64) error
	UncompleteTodo(ctx context.Context, projectID string, todoID int64) error
}

// client implements the Client interface
type client struct {
	base    *api.BaseClient
	project project.Client
}

// NewClient returns a new todo client
func NewClient(base *api.BaseClient, projectClient project.Client) Client {
	return &client{
		base:    base,
		project: projectClient,
	}
}

// GetProjectTodoSet fetches the todo set for a project
func (c *client) GetProjectTodoSet(ctx context.Context, projectID string) (*api.TodoSet, error) {
	// Get project dock to find the todo set
	dock, err := c.project.GetProjectDock(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Find the todoset in the dock
	for _, tool := range dock {
		if tool.Name == "todoset" {
			return &api.TodoSet{
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
func (c *client) GetTodoLists(ctx context.Context, projectID string, todoSetID int64) ([]api.TodoList, error) {
	var todoLists []api.TodoList
	path := fmt.Sprintf("/buckets/%s/todosets/%d/todolists.json", projectID, todoSetID)

	// Use paginated request to get all todo lists
	pr := api.NewPaginatedRequest(c.base)
	if err := pr.GetAll(path, &todoLists); err != nil {
		return nil, fmt.Errorf("failed to fetch todo lists: %w", err)
	}

	return todoLists, nil
}

// GetTodoList fetches a single todo list by ID
func (c *client) GetTodoList(ctx context.Context, projectID string, todoListID int64) (*api.TodoList, error) {
	var todoList api.TodoList

	path := fmt.Sprintf("/buckets/%s/todolists/%d.json", projectID, todoListID)
	if err := c.base.Get(path, &todoList); err != nil {
		return nil, fmt.Errorf("failed to fetch todo list: %w", err)
	}

	return &todoList, nil
}

// GetTodos fetches all todos in a todo list
func (c *client) GetTodos(ctx context.Context, projectID string, todoListID int64) ([]api.Todo, error) {
	var todos []api.Todo
	path := fmt.Sprintf("/buckets/%s/todolists/%d/todos.json", projectID, todoListID)

	// Use paginated request to get all todos
	pr := api.NewPaginatedRequest(c.base)
	if err := pr.GetAll(path, &todos); err != nil {
		return nil, fmt.Errorf("failed to fetch todos: %w", err)
	}

	return todos, nil
}

// GetAllTodos fetches all todos in a todo list including completed ones
func (c *client) GetAllTodos(ctx context.Context, projectID string, todoListID int64) ([]api.Todo, error) {
	var allTodos []api.Todo

	// Get incomplete todos
	incompleteTodos, err := c.GetTodos(ctx, projectID, todoListID)
	if err != nil {
		return nil, err
	}
	allTodos = append(allTodos, incompleteTodos...)

	// Get completed todos using the completed=true parameter
	var completedTodos []api.Todo
	path := fmt.Sprintf("/buckets/%s/todolists/%d/todos.json?completed=true", projectID, todoListID)

	// Use paginated request to get all completed todos
	pr := api.NewPaginatedRequest(c.base)
	if err := pr.GetAll(path, &completedTodos); err != nil {
		// If we can't get completed todos, just return the incomplete ones
		return allTodos, nil
	}

	// Mark them as completed (in case the API doesn't set this)
	for i := range completedTodos {
		completedTodos[i].Completed = true
	}

	allTodos = append(allTodos, completedTodos...)
	return allTodos, nil
}

// GetTodo fetches a single todo by ID
func (c *client) GetTodo(ctx context.Context, projectID string, todoID int64) (*api.Todo, error) {
	var todo api.Todo

	path := fmt.Sprintf("/buckets/%s/todos/%d.json", projectID, todoID)
	if err := c.base.Get(path, &todo); err != nil {
		return nil, fmt.Errorf("failed to fetch todo: %w", err)
	}

	return &todo, nil
}

// GetTodoGroups fetches all groups in a todo list
func (c *client) GetTodoGroups(ctx context.Context, projectID string, todoListID int64) ([]api.TodoGroup, error) {
	var groups []api.TodoGroup
	path := fmt.Sprintf("/buckets/%s/todolists/%d/groups.json", projectID, todoListID)

	// Use paginated request to get all groups
	pr := api.NewPaginatedRequest(c.base)
	if err := pr.GetAll(path, &groups); err != nil {
		return nil, fmt.Errorf("failed to fetch todo groups: %w", err)
	}

	return groups, nil
}

// CreateTodo creates a new todo in a todo list
func (c *client) CreateTodo(ctx context.Context, projectID string, todoListID int64, req api.TodoCreateRequest) (*api.Todo, error) {
	var todo api.Todo

	path := fmt.Sprintf("/buckets/%s/todolists/%d/todos.json", projectID, todoListID)
	if err := c.base.Post(path, req, &todo); err != nil {
		return nil, fmt.Errorf("failed to create todo: %w", err)
	}

	return &todo, nil
}

// CreateTodoList creates a new todo list in a project
func (c *client) CreateTodoList(ctx context.Context, projectID string, todoSetID int64, req api.TodoListCreateRequest) (*api.TodoList, error) {
	var todoList api.TodoList

	path := fmt.Sprintf("/buckets/%s/todosets/%d/todolists.json", projectID, todoSetID)
	if err := c.base.Post(path, req, &todoList); err != nil {
		return nil, fmt.Errorf("failed to create todo list: %w", err)
	}

	return &todoList, nil
}

// CompleteTodo marks a todo as complete
func (c *client) CompleteTodo(ctx context.Context, projectID string, todoID int64) error {
	path := fmt.Sprintf("/buckets/%s/todos/%d/completion.json", projectID, todoID)
	if err := c.base.Put(path, nil, nil); err != nil {
		return fmt.Errorf("failed to complete todo: %w", err)
	}

	return nil
}

// UncompleteTodo marks a todo as incomplete
func (c *client) UncompleteTodo(ctx context.Context, projectID string, todoID int64) error {
	path := fmt.Sprintf("/buckets/%s/todos/%d/completion.json", projectID, todoID)
	if err := c.base.Delete(path); err != nil {
		return fmt.Errorf("failed to uncomplete todo: %w", err)
	}

	return nil
}