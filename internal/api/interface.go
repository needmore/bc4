package api

import "context"

// APIClient defines the interface for interacting with the Basecamp API
type APIClient interface {
	// Project methods
	GetProjects(ctx context.Context) ([]Project, error)
	GetProject(ctx context.Context, projectID string) (*Project, error)

	// Todo methods
	GetProjectTodoSet(ctx context.Context, projectID string) (*TodoSet, error)
	GetTodoLists(ctx context.Context, projectID string, todoSetID int64) ([]TodoList, error)
	GetTodoList(ctx context.Context, projectID string, todoListID int64) (*TodoList, error)
	GetTodos(ctx context.Context, projectID string, todoListID int64) ([]Todo, error)
	GetAllTodos(ctx context.Context, projectID string, todoListID int64) ([]Todo, error)
	GetTodo(ctx context.Context, projectID string, todoID int64) (*Todo, error)
	GetTodoGroups(ctx context.Context, projectID string, todoListID int64) ([]TodoGroup, error)
	CreateTodo(ctx context.Context, projectID string, todoListID int64, req TodoCreateRequest) (*Todo, error)
	CreateTodoList(ctx context.Context, projectID string, todoSetID int64, req TodoListCreateRequest) (*TodoList, error)
	CompleteTodo(ctx context.Context, projectID string, todoID int64) error
	UncompleteTodo(ctx context.Context, projectID string, todoID int64) error

	// Campfire methods
	ListCampfires(ctx context.Context, projectID string) ([]Campfire, error)
	GetCampfire(ctx context.Context, projectID string, campfireID int64) (*Campfire, error)
	GetCampfireByName(ctx context.Context, projectID string, name string) (*Campfire, error)
	GetCampfireLines(ctx context.Context, projectID string, campfireID int64, limit int) ([]CampfireLine, error)
	PostCampfireLine(ctx context.Context, projectID string, campfireID int64, content string) (*CampfireLine, error)
	DeleteCampfireLine(ctx context.Context, projectID string, campfireID int64, lineID int64) error

	// Card table methods
	GetProjectCardTable(ctx context.Context, projectID string) (*CardTable, error)
	GetCardTable(ctx context.Context, projectID string, cardTableID int64) (*CardTable, error)
	GetCardsInColumn(ctx context.Context, projectID string, columnID int64) ([]Card, error)
	GetCard(ctx context.Context, projectID string, cardID int64) (*Card, error)
	CreateCard(ctx context.Context, projectID string, columnID int64, req CardCreateRequest) (*Card, error)
	UpdateCard(ctx context.Context, projectID string, cardID int64, req CardUpdateRequest) (*Card, error)
	MoveCard(ctx context.Context, projectID string, cardID int64, columnID int64) error
	ArchiveCard(ctx context.Context, projectID string, cardID int64) error

	// Card step methods
	CreateStep(ctx context.Context, projectID string, cardID int64, req StepCreateRequest) (*Step, error)
	UpdateStep(ctx context.Context, projectID string, stepID int64, req StepUpdateRequest) (*Step, error)
	SetStepCompletion(ctx context.Context, projectID string, stepID int64, completed bool) error
	MoveStep(ctx context.Context, projectID string, cardID int64, stepID int64, position int) error
	DeleteStep(ctx context.Context, projectID string, stepID int64) error

	// People methods
	GetProjectPeople(ctx context.Context, projectID string) ([]Person, error)
	GetPerson(ctx context.Context, personID int64) (*Person, error)
}

// Ensure Client implements APIClient interface
var _ APIClient = (*Client)(nil)
