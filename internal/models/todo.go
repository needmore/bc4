package models

import "time"

type TodoList struct {
	ID             int64     `json:"id"`
	Status         string    `json:"status"`
	Position       int       `json:"position"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Completed      bool      `json:"completed"`
	CompletedRatio string    `json:"completed_ratio"`
	URL            string    `json:"url"`
	AppURL         string    `json:"app_url"`
	TodosURL       string    `json:"todos_url"`
	GroupsURL      string    `json:"groups_url"`
	AppTodosURL    string    `json:"app_todos_url"`
}

type Todo struct {
	ID              int64      `json:"id"`
	Status          string     `json:"status"`
	Position        int        `json:"position"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	Content         string     `json:"content"`
	Description     string     `json:"description"`
	Completed       bool       `json:"completed"`
	CompletedAt     *time.Time `json:"completed_at"`
	DueOn           *time.Time `json:"due_on"`
	StartsOn        *time.Time `json:"starts_on"`
	NotifyOnDueDate bool       `json:"notify_on_due_date"`
	CommentsCount   int        `json:"comments_count"`
	URL             string     `json:"url"`
	AppURL          string     `json:"app_url"`
	BookmarkURL     string     `json:"bookmark_url"`
	Creator         Person     `json:"creator"`
	Assignees       []Person   `json:"assignees"`
}

type Person struct {
	ID             int64     `json:"id"`
	AttachableSgid string    `json:"attachable_sgid"`
	Name           string    `json:"name"`
	EmailAddress   string    `json:"email_address"`
	PersonableType string    `json:"personable_type"`
	Title          string    `json:"title"`
	Bio            string    `json:"bio"`
	Location       string    `json:"location"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Admin          bool      `json:"admin"`
	Owner          bool      `json:"owner"`
	Client         bool      `json:"client"`
	TimeZone       string    `json:"time_zone"`
	AvatarURL      string    `json:"avatar_url"`
	AvatarKind     string    `json:"avatar_kind"`
	CanPing        bool      `json:"can_ping"`
}
