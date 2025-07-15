package api

import "context"

// ModularAPIClient demonstrates how the interface could be split into smaller interfaces
// This file shows the proposed modular design without breaking existing code

// ProjectClient defines project-related operations
type ProjectClient interface {
	GetProjects(ctx context.Context) ([]Project, error)
	GetProject(ctx context.Context, projectID string) (*Project, error)
}

// TodoClient defines todo-related operations
type TodoClient interface {
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
}

// CampfireClient defines campfire-related operations
type CampfireClient interface {
	ListCampfires(ctx context.Context, projectID string) ([]Campfire, error)
	GetCampfire(ctx context.Context, projectID string, campfireID int64) (*Campfire, error)
	GetCampfireByName(ctx context.Context, projectID string, name string) (*Campfire, error)
	GetCampfireLines(ctx context.Context, projectID string, campfireID int64, limit int) ([]CampfireLine, error)
	PostCampfireLine(ctx context.Context, projectID string, campfireID int64, content string) (*CampfireLine, error)
	DeleteCampfireLine(ctx context.Context, projectID string, campfireID int64, lineID int64) error
}

// CardClient defines card table-related operations
type CardClient interface {
	GetProjectCardTable(ctx context.Context, projectID string) (*CardTable, error)
	GetCardTable(ctx context.Context, projectID string, cardTableID int64) (*CardTable, error)
	GetCardsInColumn(ctx context.Context, projectID string, columnID int64) ([]Card, error)
	GetCard(ctx context.Context, projectID string, cardID int64) (*Card, error)
	CreateCard(ctx context.Context, projectID string, columnID int64, req CardCreateRequest) (*Card, error)
	UpdateCard(ctx context.Context, projectID string, cardID int64, req CardUpdateRequest) (*Card, error)
	MoveCard(ctx context.Context, projectID string, cardID int64, columnID int64) error
	ArchiveCard(ctx context.Context, projectID string, cardID int64) error
}

// StepClient defines card step-related operations
type StepClient interface {
	CreateStep(ctx context.Context, projectID string, cardID int64, req StepCreateRequest) (*Step, error)
	UpdateStep(ctx context.Context, projectID string, stepID int64, req StepUpdateRequest) (*Step, error)
	SetStepCompletion(ctx context.Context, projectID string, stepID int64, completed bool) error
	MoveStep(ctx context.Context, projectID string, cardID int64, stepID int64, position int) error
	DeleteStep(ctx context.Context, projectID string, stepID int64) error
}

// PeopleClient defines people-related operations
type PeopleClient interface {
	GetProjectPeople(ctx context.Context, projectID string) ([]Person, error)
	GetPerson(ctx context.Context, personID int64) (*Person, error)
}

// ModularClient combines all the modular interfaces
// This could be used as a drop-in replacement for APIClient
type ModularClient interface {
	ProjectClient
	TodoClient
	CampfireClient
	CardClient
	StepClient
	PeopleClient
}