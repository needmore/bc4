package mock

import (
	"context"
	"errors"
	"fmt"

	"github.com/needmore/bc4/internal/api"
)

// MockClient is a mock implementation of the APIClient interface for testing
type MockClient struct {
	// Projects
	Projects      []api.Project
	ProjectsError error
	Project       *api.Project
	ProjectError  error

	// Todos
	TodoSet             *api.TodoSet
	TodoSetError        error
	TodoLists           []api.TodoList
	TodoListsError      error
	TodoList            *api.TodoList
	TodoListError       error
	Todos               []api.Todo
	TodosError          error
	Todo                *api.Todo
	TodoError           error
	TodoGroups          []api.TodoGroup
	TodoGroupsError     error
	CreatedTodo         *api.Todo
	CreateTodoError     error
	UpdatedTodo         *api.Todo
	UpdateTodoError     error
	CreatedTodoList     *api.TodoList
	CreateTodoListError error
	UpdatedTodoList     *api.TodoList
	UpdateTodoListError error
	CompleteTodoError   error
	UncompleteTodoError error

	// Campfires
	Campfires               []api.Campfire
	CampfiresError          error
	Campfire                *api.Campfire
	CampfireError           error
	CampfireLines           []api.CampfireLine
	CampfireLinesError      error
	PostedCampfireLine      *api.CampfireLine
	PostCampfireLineError   error
	DeleteCampfireLineError error

	// Cards
	CardTable        *api.CardTable
	CardTableError   error
	Cards            []api.Card
	CardsError       error
	Card             *api.Card
	CardError        error
	CreatedCard      *api.Card
	CreateCardError  error
	UpdatedCard      *api.Card
	UpdateCardError  error
	MoveCardError    error
	ArchiveCardError error

	// Steps
	CreatedStep            *api.Step
	CreateStepError        error
	UpdatedStep            *api.Step
	UpdateStepError        error
	SetStepCompletionError error
	MoveStepError          error
	DeleteStepError        error

	// People
	People       []api.Person
	PeopleError  error
	Person       *api.Person
	PersonError  error
	Profile      *api.Person
	ProfileError error

	// Activity
	Events          []api.Event
	EventsError     error
	Recordings      []api.Recording
	RecordingsError error
	Recording       *api.Recording
	RecordingError  error

	// Search
	SearchResults []api.SearchResult
	SearchError   error

	// Track method calls
	Calls []string
}

// NewMockClient creates a new mock API client
func NewMockClient() *MockClient {
	return &MockClient{
		Calls: make([]string, 0),
	}
}

// GetProjects mock implementation
func (m *MockClient) GetProjects(ctx context.Context) ([]api.Project, error) {
	m.Calls = append(m.Calls, "GetProjects")
	if m.ProjectsError != nil {
		return nil, m.ProjectsError
	}
	return m.Projects, nil
}

// GetProject mock implementation
func (m *MockClient) GetProject(ctx context.Context, projectID string) (*api.Project, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("GetProject(%s)", projectID))
	if m.ProjectError != nil {
		return nil, m.ProjectError
	}
	if m.Project == nil {
		return nil, errors.New("project not found")
	}
	return m.Project, nil
}

// GetProjectTodoSet mock implementation
func (m *MockClient) GetProjectTodoSet(ctx context.Context, projectID string) (*api.TodoSet, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("GetProjectTodoSet(%s)", projectID))
	if m.TodoSetError != nil {
		return nil, m.TodoSetError
	}
	return m.TodoSet, nil
}

// GetTodoLists mock implementation
func (m *MockClient) GetTodoLists(ctx context.Context, projectID string, todoSetID int64) ([]api.TodoList, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("GetTodoLists(%s, %d)", projectID, todoSetID))
	if m.TodoListsError != nil {
		return nil, m.TodoListsError
	}
	return m.TodoLists, nil
}

// GetTodoList mock implementation
func (m *MockClient) GetTodoList(ctx context.Context, projectID string, todoListID int64) (*api.TodoList, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("GetTodoList(%s, %d)", projectID, todoListID))
	if m.TodoListError != nil {
		return nil, m.TodoListError
	}
	return m.TodoList, nil
}

// GetTodos mock implementation
func (m *MockClient) GetTodos(ctx context.Context, projectID string, todoListID int64) ([]api.Todo, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("GetTodos(%s, %d)", projectID, todoListID))
	if m.TodosError != nil {
		return nil, m.TodosError
	}
	return m.Todos, nil
}

// GetAllTodos mock implementation
func (m *MockClient) GetAllTodos(ctx context.Context, projectID string, todoListID int64) ([]api.Todo, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("GetAllTodos(%s, %d)", projectID, todoListID))
	if m.TodosError != nil {
		return nil, m.TodosError
	}
	return m.Todos, nil
}

// GetTodo mock implementation
func (m *MockClient) GetTodo(ctx context.Context, projectID string, todoID int64) (*api.Todo, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("GetTodo(%s, %d)", projectID, todoID))
	if m.TodoError != nil {
		return nil, m.TodoError
	}
	return m.Todo, nil
}

// GetTodoGroups mock implementation
func (m *MockClient) GetTodoGroups(ctx context.Context, projectID string, todoListID int64) ([]api.TodoGroup, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("GetTodoGroups(%s, %d)", projectID, todoListID))
	if m.TodoGroupsError != nil {
		return nil, m.TodoGroupsError
	}
	return m.TodoGroups, nil
}

// CreateTodo mock implementation
func (m *MockClient) CreateTodo(ctx context.Context, projectID string, todoListID int64, req api.TodoCreateRequest) (*api.Todo, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("CreateTodo(%s, %d, %+v)", projectID, todoListID, req))
	if m.CreateTodoError != nil {
		return nil, m.CreateTodoError
	}
	return m.CreatedTodo, nil
}

// UpdateTodo mock implementation
func (m *MockClient) UpdateTodo(ctx context.Context, projectID string, todoID int64, req api.TodoUpdateRequest) (*api.Todo, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("UpdateTodo(%s, %d, %+v)", projectID, todoID, req))
	if m.UpdateTodoError != nil {
		return nil, m.UpdateTodoError
	}
	return m.UpdatedTodo, nil
}

// CreateTodoList mock implementation
func (m *MockClient) CreateTodoList(ctx context.Context, projectID string, todoSetID int64, req api.TodoListCreateRequest) (*api.TodoList, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("CreateTodoList(%s, %d, %+v)", projectID, todoSetID, req))
	if m.CreateTodoListError != nil {
		return nil, m.CreateTodoListError
	}
	return m.CreatedTodoList, nil
}

// UpdateTodoList mock implementation
func (m *MockClient) UpdateTodoList(ctx context.Context, projectID string, todoListID int64, req api.TodoListUpdateRequest) (*api.TodoList, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("UpdateTodoList(%s, %d, %+v)", projectID, todoListID, req))
	if m.UpdateTodoListError != nil {
		return nil, m.UpdateTodoListError
	}
	return m.UpdatedTodoList, nil
}

// CreateTodoGroup mock implementation
func (m *MockClient) CreateTodoGroup(ctx context.Context, projectID string, todoListID int64, req api.TodoGroupCreateRequest) (*api.TodoGroup, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("CreateTodoGroup(%s, %d, %+v)", projectID, todoListID, req))
	return &api.TodoGroup{ID: 1, Name: req.Name}, nil
}

// RepositionTodoGroup mock implementation
func (m *MockClient) RepositionTodoGroup(ctx context.Context, projectID string, groupID int64, position int) error {
	m.Calls = append(m.Calls, fmt.Sprintf("RepositionTodoGroup(%s, %d, %d)", projectID, groupID, position))
	return nil
}

// RepositionTodo mock implementation
func (m *MockClient) RepositionTodo(ctx context.Context, projectID string, todoID int64, position int) error {
	m.Calls = append(m.Calls, fmt.Sprintf("RepositionTodo(%s, %d, %d)", projectID, todoID, position))
	return nil
}

// CompleteTodo mock implementation
func (m *MockClient) CompleteTodo(ctx context.Context, projectID string, todoID int64) error {
	m.Calls = append(m.Calls, fmt.Sprintf("CompleteTodo(%s, %d)", projectID, todoID))
	return m.CompleteTodoError
}

// UncompleteTodo mock implementation
func (m *MockClient) UncompleteTodo(ctx context.Context, projectID string, todoID int64) error {
	m.Calls = append(m.Calls, fmt.Sprintf("UncompleteTodo(%s, %d)", projectID, todoID))
	return m.UncompleteTodoError
}

// ListCampfires mock implementation
func (m *MockClient) ListCampfires(ctx context.Context, projectID string) ([]api.Campfire, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("ListCampfires(%s)", projectID))
	if m.CampfiresError != nil {
		return nil, m.CampfiresError
	}
	return m.Campfires, nil
}

// GetCampfire mock implementation
func (m *MockClient) GetCampfire(ctx context.Context, projectID string, campfireID int64) (*api.Campfire, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("GetCampfire(%s, %d)", projectID, campfireID))
	if m.CampfireError != nil {
		return nil, m.CampfireError
	}
	return m.Campfire, nil
}

// GetCampfireByName mock implementation
func (m *MockClient) GetCampfireByName(ctx context.Context, projectID string, name string) (*api.Campfire, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("GetCampfireByName(%s, %s)", projectID, name))
	if m.CampfireError != nil {
		return nil, m.CampfireError
	}
	return m.Campfire, nil
}

// GetCampfireLines mock implementation
func (m *MockClient) GetCampfireLines(ctx context.Context, projectID string, campfireID int64, limit int) ([]api.CampfireLine, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("GetCampfireLines(%s, %d, %d)", projectID, campfireID, limit))
	if m.CampfireLinesError != nil {
		return nil, m.CampfireLinesError
	}
	return m.CampfireLines, nil
}

// PostCampfireLine mock implementation
func (m *MockClient) PostCampfireLine(ctx context.Context, projectID string, campfireID int64, content string, contentType string) (*api.CampfireLine, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("PostCampfireLine(%s, %d, %s, %s)", projectID, campfireID, content, contentType))
	if m.PostCampfireLineError != nil {
		return nil, m.PostCampfireLineError
	}
	return m.PostedCampfireLine, nil
}

// DeleteCampfireLine mock implementation
func (m *MockClient) DeleteCampfireLine(ctx context.Context, projectID string, campfireID int64, lineID int64) error {
	m.Calls = append(m.Calls, fmt.Sprintf("DeleteCampfireLine(%s, %d, %d)", projectID, campfireID, lineID))
	return m.DeleteCampfireLineError
}

// GetAllProjectCardTables mock implementation
func (m *MockClient) GetAllProjectCardTables(ctx context.Context, projectID string) ([]*api.CardTable, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("GetAllProjectCardTables(%s)", projectID))
	if m.CardTableError != nil {
		return nil, m.CardTableError
	}
	if m.CardTable != nil {
		return []*api.CardTable{m.CardTable}, nil
	}
	return []*api.CardTable{}, nil
}

// GetProjectCardTable mock implementation
func (m *MockClient) GetProjectCardTable(ctx context.Context, projectID string) (*api.CardTable, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("GetProjectCardTable(%s)", projectID))
	if m.CardTableError != nil {
		return nil, m.CardTableError
	}
	return m.CardTable, nil
}

// GetCardTable mock implementation
func (m *MockClient) GetCardTable(ctx context.Context, projectID string, cardTableID int64) (*api.CardTable, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("GetCardTable(%s, %d)", projectID, cardTableID))
	if m.CardTableError != nil {
		return nil, m.CardTableError
	}
	return m.CardTable, nil
}

// GetCardsInColumn mock implementation
func (m *MockClient) GetCardsInColumn(ctx context.Context, projectID string, columnID int64) ([]api.Card, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("GetCardsInColumn(%s, %d)", projectID, columnID))
	if m.CardsError != nil {
		return nil, m.CardsError
	}
	return m.Cards, nil
}

// GetCard mock implementation
func (m *MockClient) GetCard(ctx context.Context, projectID string, cardID int64) (*api.Card, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("GetCard(%s, %d)", projectID, cardID))
	if m.CardError != nil {
		return nil, m.CardError
	}
	return m.Card, nil
}

// CreateCard mock implementation
func (m *MockClient) CreateCard(ctx context.Context, projectID string, columnID int64, req api.CardCreateRequest) (*api.Card, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("CreateCard(%s, %d, %+v)", projectID, columnID, req))
	if m.CreateCardError != nil {
		return nil, m.CreateCardError
	}
	return m.CreatedCard, nil
}

// UpdateCard mock implementation
func (m *MockClient) UpdateCard(ctx context.Context, projectID string, cardID int64, req api.CardUpdateRequest) (*api.Card, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("UpdateCard(%s, %d, %+v)", projectID, cardID, req))
	if m.UpdateCardError != nil {
		return nil, m.UpdateCardError
	}
	return m.UpdatedCard, nil
}

// MoveCard mock implementation
func (m *MockClient) MoveCard(ctx context.Context, projectID string, cardID int64, columnID int64) error {
	m.Calls = append(m.Calls, fmt.Sprintf("MoveCard(%s, %d, %d)", projectID, cardID, columnID))
	return m.MoveCardError
}

// ArchiveCard mock implementation
func (m *MockClient) ArchiveCard(ctx context.Context, projectID string, cardID int64) error {
	m.Calls = append(m.Calls, fmt.Sprintf("ArchiveCard(%s, %d)", projectID, cardID))
	return m.ArchiveCardError
}

// CreateStep mock implementation
func (m *MockClient) CreateStep(ctx context.Context, projectID string, cardID int64, req api.StepCreateRequest) (*api.Step, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("CreateStep(%s, %d, %+v)", projectID, cardID, req))
	if m.CreateStepError != nil {
		return nil, m.CreateStepError
	}
	return m.CreatedStep, nil
}

// UpdateStep mock implementation
func (m *MockClient) UpdateStep(ctx context.Context, projectID string, stepID int64, req api.StepUpdateRequest) (*api.Step, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("UpdateStep(%s, %d, %+v)", projectID, stepID, req))
	if m.UpdateStepError != nil {
		return nil, m.UpdateStepError
	}
	return m.UpdatedStep, nil
}

// SetStepCompletion mock implementation
func (m *MockClient) SetStepCompletion(ctx context.Context, projectID string, stepID int64, completed bool) error {
	m.Calls = append(m.Calls, fmt.Sprintf("SetStepCompletion(%s, %d, %v)", projectID, stepID, completed))
	return m.SetStepCompletionError
}

// MoveStep mock implementation
func (m *MockClient) MoveStep(ctx context.Context, projectID string, cardID int64, stepID int64, position int) error {
	m.Calls = append(m.Calls, fmt.Sprintf("MoveStep(%s, %d, %d, %d)", projectID, cardID, stepID, position))
	return m.MoveStepError
}

// DeleteStep mock implementation
func (m *MockClient) DeleteStep(ctx context.Context, projectID string, stepID int64) error {
	m.Calls = append(m.Calls, fmt.Sprintf("DeleteStep(%s, %d)", projectID, stepID))
	return m.DeleteStepError
}

// GetProjectPeople mock implementation
func (m *MockClient) GetProjectPeople(ctx context.Context, projectID string) ([]api.Person, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("GetProjectPeople(%s)", projectID))
	if m.PeopleError != nil {
		return nil, m.PeopleError
	}
	return m.People, nil
}

// GetPerson mock implementation
func (m *MockClient) GetPerson(ctx context.Context, personID int64) (*api.Person, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("GetPerson(%d)", personID))
	if m.PersonError != nil {
		return nil, m.PersonError
	}
	return m.Person, nil
}

// GetMyProfile mock implementation
func (m *MockClient) GetMyProfile(ctx context.Context) (*api.Person, error) {
	m.Calls = append(m.Calls, "GetMyProfile()")
	if m.ProfileError != nil {
		return nil, m.ProfileError
	}
	return m.Profile, nil
}

// GetAllPeople mock implementation
func (m *MockClient) GetAllPeople(ctx context.Context) ([]api.Person, error) {
	m.Calls = append(m.Calls, "GetAllPeople()")
	if m.PeopleError != nil {
		return nil, m.PeopleError
	}
	return m.People, nil
}

// GetPingablePeople mock implementation
func (m *MockClient) GetPingablePeople(ctx context.Context) ([]api.Person, error) {
	m.Calls = append(m.Calls, "GetPingablePeople()")
	if m.PeopleError != nil {
		return nil, m.PeopleError
	}
	return m.People, nil
}

// UpdateProjectAccess mock implementation
func (m *MockClient) UpdateProjectAccess(ctx context.Context, projectID string, req api.ProjectAccessUpdateRequest) (*api.ProjectAccessUpdateResponse, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("UpdateProjectAccess(%s, %+v)", projectID, req))
	if m.PeopleError != nil {
		return nil, m.PeopleError
	}
	return &api.ProjectAccessUpdateResponse{
		Granted: m.People,
		Revoked: []api.Person{},
	}, nil
}

// ListEvents mock implementation
func (m *MockClient) ListEvents(ctx context.Context, projectID string, recordingID int64) ([]api.Event, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("ListEvents(%s, %d)", projectID, recordingID))
	if m.EventsError != nil {
		return nil, m.EventsError
	}
	return m.Events, nil
}

// ListRecordings mock implementation
func (m *MockClient) ListRecordings(ctx context.Context, projectID string, opts *api.ActivityListOptions) ([]api.Recording, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("ListRecordings(%s, %+v)", projectID, opts))
	if m.RecordingsError != nil {
		return nil, m.RecordingsError
	}
	return m.Recordings, nil
}

// GetRecording mock implementation
func (m *MockClient) GetRecording(ctx context.Context, projectID string, recordingID int64) (*api.Recording, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("GetRecording(%s, %d)", projectID, recordingID))
	if m.RecordingError != nil {
		return nil, m.RecordingError
	}
	return m.Recording, nil
}

// GetProjectSchedule mock implementation
func (m *MockClient) GetProjectSchedule(ctx context.Context, projectID string) (*api.Schedule, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("GetProjectSchedule(%s)", projectID))
	return &api.Schedule{ID: 1, Title: "Schedule"}, nil
}

// GetSchedule mock implementation
func (m *MockClient) GetSchedule(ctx context.Context, projectID string, scheduleID int64) (*api.Schedule, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("GetSchedule(%s, %d)", projectID, scheduleID))
	return &api.Schedule{ID: scheduleID, Title: "Schedule"}, nil
}

// GetScheduleEntries mock implementation
func (m *MockClient) GetScheduleEntries(ctx context.Context, projectID string, scheduleID int64) ([]api.ScheduleEntry, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("GetScheduleEntries(%s, %d)", projectID, scheduleID))
	return []api.ScheduleEntry{}, nil
}

// GetScheduleEntriesInRange mock implementation
func (m *MockClient) GetScheduleEntriesInRange(ctx context.Context, projectID string, scheduleID int64, startDate, endDate string) ([]api.ScheduleEntry, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("GetScheduleEntriesInRange(%s, %d, %s, %s)", projectID, scheduleID, startDate, endDate))
	return []api.ScheduleEntry{}, nil
}

// GetUpcomingScheduleEntries mock implementation
func (m *MockClient) GetUpcomingScheduleEntries(ctx context.Context, projectID string, scheduleID int64) ([]api.ScheduleEntry, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("GetUpcomingScheduleEntries(%s, %d)", projectID, scheduleID))
	return []api.ScheduleEntry{}, nil
}

// GetPastScheduleEntries mock implementation
func (m *MockClient) GetPastScheduleEntries(ctx context.Context, projectID string, scheduleID int64) ([]api.ScheduleEntry, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("GetPastScheduleEntries(%s, %d)", projectID, scheduleID))
	return []api.ScheduleEntry{}, nil
}

// GetScheduleEntry mock implementation
func (m *MockClient) GetScheduleEntry(ctx context.Context, projectID string, entryID int64) (*api.ScheduleEntry, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("GetScheduleEntry(%s, %d)", projectID, entryID))
	return &api.ScheduleEntry{ID: entryID, Title: "Event"}, nil
}

// CreateScheduleEntry mock implementation
func (m *MockClient) CreateScheduleEntry(ctx context.Context, projectID string, scheduleID int64, req api.ScheduleEntryCreateRequest) (*api.ScheduleEntry, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("CreateScheduleEntry(%s, %d, %+v)", projectID, scheduleID, req))
	return &api.ScheduleEntry{ID: 1, Title: req.Summary}, nil
}

// UpdateScheduleEntry mock implementation
func (m *MockClient) UpdateScheduleEntry(ctx context.Context, projectID string, entryID int64, req api.ScheduleEntryUpdateRequest) (*api.ScheduleEntry, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("UpdateScheduleEntry(%s, %d, %+v)", projectID, entryID, req))
	return &api.ScheduleEntry{ID: entryID}, nil
}

// DeleteScheduleEntry mock implementation
func (m *MockClient) DeleteScheduleEntry(ctx context.Context, projectID string, entryID int64) error {
	m.Calls = append(m.Calls, fmt.Sprintf("DeleteScheduleEntry(%s, %d)", projectID, entryID))
	return nil
}

// Search mock implementation
func (m *MockClient) Search(ctx context.Context, opts api.SearchOptions) ([]api.SearchResult, error) {
	m.Calls = append(m.Calls, fmt.Sprintf("Search(%+v)", opts))
	if m.SearchError != nil {
		return nil, m.SearchError
	}
	return m.SearchResults, nil
}

// Ensure MockClient implements APIClient interface
var _ api.APIClient = (*MockClient)(nil)
