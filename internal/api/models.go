package api

import "time"

// Project represents a Basecamp project
type Project struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Purpose     string `json:"purpose"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// DockItem represents a tool/feature in a project's dock
type DockItem struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Name  string `json:"name"`
	URL   string `json:"url"`
}

// Person represents a Basecamp user
type Person struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	EmailAddress string `json:"email_address"`
	Title        string `json:"title"`
	AvatarURL    string `json:"avatar_url"`
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

// TodoCreateRequest represents the payload for creating a new todo
type TodoCreateRequest struct {
	Content     string  `json:"content"`
	Description string  `json:"description,omitempty"`
	DueOn       *string `json:"due_on,omitempty"`
	StartsOn    *string `json:"starts_on,omitempty"`
	AssigneeIDs []int64 `json:"assignee_ids,omitempty"`
}

// TodoListCreateRequest represents the payload for creating a new todo list
type TodoListCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// Campfire represents a Basecamp campfire (chat room)
type Campfire struct {
	ID        int64     `json:"id"`
	Name      string    `json:"title"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	LinesURL  string    `json:"lines_url"`
	URL       string    `json:"url"`
	Bucket    struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"bucket"`
	Creator Person `json:"creator"`
}

// CampfireLine represents a message in a campfire
type CampfireLine struct {
	ID        int64     `json:"id"`
	Status    string    `json:"status"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	URL       string    `json:"url"`
	Creator   Person    `json:"creator"`
	Parent    struct {
		ID    int64  `json:"id"`
		Title string `json:"title"`
		Type  string `json:"type"`
		URL   string `json:"url"`
	} `json:"parent"`
}

// CampfireLineCreate represents the request body for creating a campfire line
type CampfireLineCreate struct {
	Content string `json:"content"`
}

// CardTable represents a Basecamp card table (kanban board)
type CardTable struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Lists       []Column  `json:"lists"`
	CardsCount  int       `json:"cards_count"`
	URL         string    `json:"url"`
}

// Column represents a column in a card table
type Column struct {
	ID         int64     `json:"id"`
	Title      string    `json:"title"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	Color      string    `json:"color,omitempty"`
	Status     string    `json:"status"`
	OnHold     bool      `json:"on_hold"`
	CardsCount int       `json:"cards_count"`
	CardsURL   string    `json:"cards_url"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Card represents a card in a card table
type Card struct {
	ID         int64     `json:"id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	Status     string    `json:"status"`
	DueOn      *string   `json:"due_on,omitempty"`
	Assignees  []Person  `json:"assignees"`
	Steps      []Step    `json:"steps"`
	StepsCount int       `json:"steps_count"`
	Creator    *Person   `json:"creator"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Parent     *Column   `json:"parent"`
	URL        string    `json:"url"`
}

// Step represents a step within a card
type Step struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	Completed bool      `json:"completed"`
	DueOn     *string   `json:"due_on,omitempty"`
	Assignees []Person  `json:"assignees"`
	Creator   *Person   `json:"creator"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CardCreateRequest represents the payload for creating a new card
type CardCreateRequest struct {
	Title   string  `json:"title"`
	Content string  `json:"content,omitempty"`
	DueOn   *string `json:"due_on,omitempty"`
	Notify  bool    `json:"notify,omitempty"`
}

// CardUpdateRequest represents the payload for updating a card
type CardUpdateRequest struct {
	Title       string  `json:"title,omitempty"`
	Content     string  `json:"content,omitempty"`
	DueOn       *string `json:"due_on,omitempty"`
	AssigneeIDs []int64 `json:"assignee_ids,omitempty"`
}

// StepCreateRequest represents the payload for creating a new step
type StepCreateRequest struct {
	Title     string  `json:"title"`
	DueOn     *string `json:"due_on,omitempty"`
	Assignees string  `json:"assignees,omitempty"` // comma-separated list of person IDs
}

// StepUpdateRequest represents the payload for updating a step
type StepUpdateRequest struct {
	Title     string  `json:"title,omitempty"`
	DueOn     *string `json:"due_on,omitempty"`
	Assignees string  `json:"assignees,omitempty"`
}