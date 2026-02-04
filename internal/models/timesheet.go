package models

import "time"

// TimesheetEntry represents a single timesheet entry in Basecamp
type TimesheetEntry struct {
	ID          int64     `json:"id"`
	Status      string    `json:"status"`
	Type        string    `json:"type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Date        string    `json:"date"`        // ISO 8601 date (YYYY-MM-DD)
	Description string    `json:"description"` // Work description
	Hours       float64   `json:"hours"`       // Hours logged
	Creator     Person    `json:"creator"`     // Person who created the entry
	Parent      Parent    `json:"parent"`      // Associated recording/project
	Bucket      Bucket    `json:"bucket"`      // Project bucket
	URL         string    `json:"url"`
	AppURL      string    `json:"app_url"`
	BookmarkURL string    `json:"bookmark_url"`
}

// Parent represents the parent recording for a timesheet entry
type Parent struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	Type   string `json:"type"`
	URL    string `json:"url"`
	AppURL string `json:"app_url"`
}

// Bucket represents the project bucket for a timesheet entry
type Bucket struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}
