package api

import (
	"context"
)

// ProjectOperations defines project-specific operations
type ProjectOperations interface {
	GetProjects(ctx context.Context) ([]Project, error)
	GetProject(ctx context.Context, projectID string) (*Project, error)
	CreateProject(ctx context.Context, req ProjectCreateRequest) (*Project, error)
	UpdateProject(ctx context.Context, projectID string, req ProjectUpdateRequest) (*Project, error)
	DeleteProject(ctx context.Context, projectID string) error
	ArchiveProject(ctx context.Context, projectID string) error
	UnarchiveProject(ctx context.Context, projectID string) error
	CopyProject(ctx context.Context, sourceProjectID string, name string, description string) (*Project, error)
}

// TodoOperations defines todo-specific operations
type TodoOperations interface {
	GetProjectTodoSet(ctx context.Context, projectID string) (*TodoSet, error)
	GetTodoLists(ctx context.Context, projectID string, todoSetID int64) ([]TodoList, error)
	GetTodoList(ctx context.Context, projectID string, todoListID int64) (*TodoList, error)
	GetTodos(ctx context.Context, projectID string, todoListID int64) ([]Todo, error)
	GetAllTodos(ctx context.Context, projectID string, todoListID int64) ([]Todo, error)
	GetTodo(ctx context.Context, projectID string, todoID int64) (*Todo, error)
	GetTodoGroups(ctx context.Context, projectID string, todoListID int64) ([]TodoGroup, error)
	CreateTodo(ctx context.Context, projectID string, todoListID int64, req TodoCreateRequest) (*Todo, error)
	UpdateTodo(ctx context.Context, projectID string, todoID int64, req TodoUpdateRequest) (*Todo, error)
	CreateTodoList(ctx context.Context, projectID string, todoSetID int64, req TodoListCreateRequest) (*TodoList, error)
	UpdateTodoList(ctx context.Context, projectID string, todoListID int64, req TodoListUpdateRequest) (*TodoList, error)
	CreateTodoGroup(ctx context.Context, projectID string, todoListID int64, req TodoGroupCreateRequest) (*TodoGroup, error)
	RepositionTodoGroup(ctx context.Context, projectID string, groupID int64, position int) error
	RepositionTodo(ctx context.Context, projectID string, todoID int64, position int) error
	CompleteTodo(ctx context.Context, projectID string, todoID int64) error
	UncompleteTodo(ctx context.Context, projectID string, todoID int64) error
}

// CampfireOperations defines campfire-specific operations
type CampfireOperations interface {
	ListCampfires(ctx context.Context, projectID string) ([]Campfire, error)
	GetCampfire(ctx context.Context, projectID string, campfireID int64) (*Campfire, error)
	GetCampfireByName(ctx context.Context, projectID string, name string) (*Campfire, error)
	GetCampfireLines(ctx context.Context, projectID string, campfireID int64, limit int) ([]CampfireLine, error)
	PostCampfireLine(ctx context.Context, projectID string, campfireID int64, content string, contentType string) (*CampfireLine, error)
	DeleteCampfireLine(ctx context.Context, projectID string, campfireID int64, lineID int64) error
}

// CardOperations defines card table-specific operations
type CardOperations interface {
	GetAllProjectCardTables(ctx context.Context, projectID string) ([]*CardTable, error)
	GetProjectCardTable(ctx context.Context, projectID string) (*CardTable, error)
	GetCardTable(ctx context.Context, projectID string, cardTableID int64) (*CardTable, error)
	GetCardsInColumn(ctx context.Context, projectID string, columnID int64) ([]Card, error)
	GetOnHoldCardsInColumn(ctx context.Context, onHoldCardsURL string) ([]Card, error)
	GetCard(ctx context.Context, projectID string, cardID int64) (*Card, error)
	CreateCard(ctx context.Context, projectID string, columnID int64, req CardCreateRequest) (*Card, error)
	UpdateCard(ctx context.Context, projectID string, cardID int64, req CardUpdateRequest) (*Card, error)
	MoveCard(ctx context.Context, projectID string, cardID int64, columnID int64) error
	ArchiveCard(ctx context.Context, projectID string, cardID int64) error
}

// StepOperations defines card step-specific operations
type StepOperations interface {
	CreateStep(ctx context.Context, projectID string, cardID int64, req StepCreateRequest) (*Step, error)
	UpdateStep(ctx context.Context, projectID string, stepID int64, req StepUpdateRequest) (*Step, error)
	SetStepCompletion(ctx context.Context, projectID string, stepID int64, completed bool) error
	MoveStep(ctx context.Context, projectID string, cardID int64, stepID int64, position int) error
	DeleteStep(ctx context.Context, projectID string, stepID int64) error
}

// ColumnOperations defines column-specific operations
type ColumnOperations interface {
	CreateColumn(ctx context.Context, projectID string, cardTableID int64, req ColumnCreateRequest) (*Column, error)
	UpdateColumn(ctx context.Context, projectID string, columnID int64, req ColumnUpdateRequest) (*Column, error)
	SetColumnColor(ctx context.Context, projectID string, columnID int64, color string) error
	MoveColumn(ctx context.Context, projectID string, cardTableID int64, sourceID, targetID int64, position string) error
	SetColumnOnHold(ctx context.Context, projectID string, columnID int64) error
	RemoveColumnOnHold(ctx context.Context, projectID string, columnID int64) error
}

// PeopleOperations defines people-specific operations
type PeopleOperations interface {
	GetProjectPeople(ctx context.Context, projectID string) ([]Person, error)
	GetPerson(ctx context.Context, personID int64) (*Person, error)
	GetMyProfile(ctx context.Context) (*Person, error)
	GetAllPeople(ctx context.Context) ([]Person, error)
	GetPingablePeople(ctx context.Context) ([]Person, error)
	UpdateProjectAccess(ctx context.Context, projectID string, req ProjectAccessUpdateRequest) (*ProjectAccessUpdateResponse, error)
}

// AttachmentOperations defines attachment-specific operations
type AttachmentOperations interface {
	UploadAttachment(filename string, data []byte, contentType string) (*AttachmentUploadResponse, error)
}

// UploadOperations defines upload-specific operations
type UploadOperations interface {
	GetUpload(ctx context.Context, bucketID string, uploadID int64) (*Upload, error)
	DownloadAttachment(ctx context.Context, downloadURL, destPath string) error
}

// CommentOperations defines comment-specific operations
type CommentOperations interface {
	ListComments(ctx context.Context, projectID string, recordingID int64) ([]Comment, error)
	GetComment(ctx context.Context, projectID string, commentID int64) (*Comment, error)
	CreateComment(ctx context.Context, projectID string, recordingID int64, req CommentCreateRequest) (*Comment, error)
	UpdateComment(ctx context.Context, projectID string, commentID int64, req CommentUpdateRequest) (*Comment, error)
}

// ActivityOperations defines activity-specific operations
type ActivityOperations interface {
	ListEvents(ctx context.Context, projectID string, recordingID int64) ([]Event, error)
	ListRecordings(ctx context.Context, projectID string, opts *ActivityListOptions) ([]Recording, error)
	GetRecording(ctx context.Context, projectID string, recordingID int64) (*Recording, error)
}

// ScheduleOperations defines schedule-specific operations
type ScheduleOperations interface {
	GetProjectSchedule(ctx context.Context, projectID string) (*Schedule, error)
	GetSchedule(ctx context.Context, projectID string, scheduleID int64) (*Schedule, error)
	GetScheduleEntries(ctx context.Context, projectID string, scheduleID int64) ([]ScheduleEntry, error)
	GetScheduleEntriesInRange(ctx context.Context, projectID string, scheduleID int64, startDate, endDate string) ([]ScheduleEntry, error)
	GetUpcomingScheduleEntries(ctx context.Context, projectID string, scheduleID int64) ([]ScheduleEntry, error)
	GetPastScheduleEntries(ctx context.Context, projectID string, scheduleID int64) ([]ScheduleEntry, error)
	GetScheduleEntry(ctx context.Context, projectID string, entryID int64) (*ScheduleEntry, error)
	CreateScheduleEntry(ctx context.Context, projectID string, scheduleID int64, req ScheduleEntryCreateRequest) (*ScheduleEntry, error)
	UpdateScheduleEntry(ctx context.Context, projectID string, entryID int64, req ScheduleEntryUpdateRequest) (*ScheduleEntry, error)
	DeleteScheduleEntry(ctx context.Context, projectID string, entryID int64) error
}

// ModularClient provides access to all API operations through focused interfaces
type ModularClient struct {
	*Client // Embed the existing client for now
}

// NewModularClient creates a new modular client that exposes focused interfaces
func NewModularClient(accountID, accessToken string) *ModularClient {
	return &ModularClient{
		Client: NewClient(accountID, accessToken),
	}
}

// Projects returns the project operations interface
func (c *ModularClient) Projects() ProjectOperations {
	return c.Client
}

// Todos returns the todo operations interface
func (c *ModularClient) Todos() TodoOperations {
	return c.Client
}

// Campfires returns the campfire operations interface
func (c *ModularClient) Campfires() CampfireOperations {
	return c.Client
}

// Cards returns the card operations interface
func (c *ModularClient) Cards() CardOperations {
	return c.Client
}

// Steps returns the step operations interface
func (c *ModularClient) Steps() StepOperations {
	return c.Client
}

// Columns returns the column operations interface
func (c *ModularClient) Columns() ColumnOperations {
	return c.Client
}

// People returns the people operations interface
func (c *ModularClient) People() PeopleOperations {
	return c.Client
}

// Attachments returns the attachment operations interface
func (c *ModularClient) Attachments() AttachmentOperations {
	return c.Client
}

// Comments returns the comment operations interface
func (c *ModularClient) Comments() CommentOperations {
	return c.Client
}

// Activity returns the activity operations interface
func (c *ModularClient) Activity() ActivityOperations {
	return c.Client
}

// Schedules returns the schedule operations interface
func (c *ModularClient) Schedules() ScheduleOperations {
	return c.Client
}

// Uploads returns the upload operations interface
func (c *ModularClient) Uploads() UploadOperations {
	return c.Client
}

// Search returns the search operations interface
func (c *ModularClient) Search() SearchOperations {
	return c.Client
}

// QuestionOperations defines check-in question operations
type QuestionOperations interface {
	GetProjectQuestionnaire(ctx context.Context, projectID string) (*Questionnaire, error)
	ListQuestions(ctx context.Context, projectID string, questionnaireID int64) ([]Question, error)
	GetQuestion(ctx context.Context, projectID string, questionID int64) (*Question, error)
	CreateQuestion(ctx context.Context, projectID string, questionnaireID int64, req QuestionCreateRequest) (*Question, error)
	UpdateQuestion(ctx context.Context, projectID string, questionID int64, req QuestionUpdateRequest) (*Question, error)
	PauseQuestion(ctx context.Context, projectID string, questionID int64) error
	ResumeQuestion(ctx context.Context, projectID string, questionID int64) error
	UpdateNotificationSettings(ctx context.Context, projectID string, questionID int64, req NotificationSettingsUpdateRequest) (*QuestionNotificationSettings, error)
	ListAnswers(ctx context.Context, projectID string, questionID int64, opts *AnswerListOptions) ([]QuestionAnswer, error)
	ListAnswerers(ctx context.Context, projectID string, questionID int64) ([]Person, error)
	GetAnswersByPerson(ctx context.Context, projectID string, questionID int64, personID int64) ([]QuestionAnswer, error)
	GetAnswer(ctx context.Context, projectID string, answerID int64) (*QuestionAnswer, error)
	CreateAnswer(ctx context.Context, projectID string, questionID int64, req AnswerCreateRequest) (*QuestionAnswer, error)
	UpdateAnswer(ctx context.Context, projectID string, answerID int64, req AnswerUpdateRequest) (*QuestionAnswer, error)
	ListMyReminders(ctx context.Context) ([]QuestionReminder, error)
}

// Questions returns the question operations interface
func (c *ModularClient) Questions() QuestionOperations {
	return c.Client
}

// Example of how to extend with new operations without modifying existing code:
//
// type MessageOperations interface {
//     GetMessages(ctx context.Context, projectID string) ([]Message, error)
//     PostMessage(ctx context.Context, projectID string, content string) (*Message, error)
// }
//
// func (c *ModularClient) Messages() MessageOperations {
//     return c.Client
// }
