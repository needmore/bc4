package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// Schedule represents a Basecamp schedule (calendar)
type Schedule struct {
	ID               int64   `json:"id"`
	Status           string  `json:"status"`
	VisibleToClients bool    `json:"visible_to_clients"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
	Title            string  `json:"title"`
	InheritsStatus   bool    `json:"inherits_status"`
	Type             string  `json:"type"`
	URL              string  `json:"url"`
	AppURL           string  `json:"app_url"`
	BookmarkURL      string  `json:"bookmark_url"`
	Position         int     `json:"position"`
	Bucket           *Bucket `json:"bucket,omitempty"`
	Creator          *Person `json:"creator,omitempty"`
	EntriesCount     int     `json:"entries_count"`
	EntriesURL       string  `json:"entries_url"`
}

// ScheduleEntry represents a calendar event in Basecamp
type ScheduleEntry struct {
	ID               int64       `json:"id"`
	Status           string      `json:"status"`
	VisibleToClients bool        `json:"visible_to_clients"`
	CreatedAt        string      `json:"created_at"`
	UpdatedAt        string      `json:"updated_at"`
	Title            string      `json:"title"`
	InheritsStatus   bool        `json:"inherits_status"`
	Type             string      `json:"type"`
	URL              string      `json:"url"`
	AppURL           string      `json:"app_url"`
	BookmarkURL      string      `json:"bookmark_url"`
	SubscriptionURL  string      `json:"subscription_url"`
	CommentsCount    int         `json:"comments_count"`
	CommentsURL      string      `json:"comments_url"`
	Parent           *Parent     `json:"parent,omitempty"`
	Bucket           *Bucket     `json:"bucket,omitempty"`
	Creator          *Person     `json:"creator,omitempty"`
	Description      string      `json:"description"`
	Summary          string      `json:"summary"`
	AllDay           bool        `json:"all_day"`
	StartsAt         string      `json:"starts_at"`
	EndsAt           string      `json:"ends_at"`
	Participants     []Person    `json:"participants"`
	Recurrence       *Recurrence `json:"recurrence_schedule,omitempty"`
}

// Recurrence represents recurrence schedule for an event
// TODO: Add command-level support for creating recurring events
// This struct is currently only used for reading existing recurring events from the API
type Recurrence struct {
	Frequency string `json:"frequency"`      // "every_day", "every_week", "every_month", "every_year"
	Days      []int  `json:"days,omitempty"` // For weekly: [1,2,3,4,5] = Mon-Fri
	Week      string `json:"week,omitempty"` // For monthly: "first", "second", "third", "fourth", "last"
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
}

// ScheduleEntryCreateRequest represents the payload for creating a schedule entry
type ScheduleEntryCreateRequest struct {
	Summary        string  `json:"summary"`
	Description    string  `json:"description,omitempty"`
	StartsAt       string  `json:"starts_at"`
	EndsAt         string  `json:"ends_at"`
	AllDay         bool    `json:"all_day,omitempty"`
	ParticipantIDs []int64 `json:"participant_ids,omitempty"`
	Notify         bool    `json:"notify,omitempty"`
}

// ScheduleEntryUpdateRequest represents the payload for updating a schedule entry
type ScheduleEntryUpdateRequest struct {
	Summary        *string `json:"summary,omitempty"`
	Description    *string `json:"description,omitempty"`
	StartsAt       *string `json:"starts_at,omitempty"`
	EndsAt         *string `json:"ends_at,omitempty"`
	AllDay         *bool   `json:"all_day,omitempty"`
	ParticipantIDs []int64 `json:"participant_ids,omitempty"`
}

// GetProjectSchedule fetches the schedule (calendar) for a project from its dock
func (c *Client) GetProjectSchedule(ctx context.Context, projectID string) (*Schedule, error) {
	// Get the project to access its dock
	path := fmt.Sprintf("/projects/%s.json", projectID)

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch project: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var projectData struct {
		Dock []struct {
			ID      int64  `json:"id"`
			Title   string `json:"title"`
			Name    string `json:"name"`
			URL     string `json:"url"`
			Enabled bool   `json:"enabled"`
		} `json:"dock"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&projectData); err != nil {
		return nil, fmt.Errorf("failed to decode project data: %w", err)
	}

	// Find the schedule in the dock
	for _, tool := range projectData.Dock {
		if tool.Name == "schedule" {
			return &Schedule{
				ID:         tool.ID,
				Title:      tool.Title,
				EntriesURL: tool.URL,
			}, nil
		}
	}

	return nil, fmt.Errorf("schedule not found for project")
}

// GetSchedule fetches a specific schedule by ID
func (c *Client) GetSchedule(ctx context.Context, projectID string, scheduleID int64) (*Schedule, error) {
	var schedule Schedule
	path := fmt.Sprintf("/buckets/%s/schedules/%d.json", projectID, scheduleID)

	if err := c.Get(path, &schedule); err != nil {
		return nil, fmt.Errorf("failed to fetch schedule: %w", err)
	}

	return &schedule, nil
}

// GetScheduleEntries fetches all entries in a schedule
func (c *Client) GetScheduleEntries(ctx context.Context, projectID string, scheduleID int64) ([]ScheduleEntry, error) {
	var entries []ScheduleEntry
	path := fmt.Sprintf("/buckets/%s/schedules/%d/entries.json", projectID, scheduleID)

	pr := NewPaginatedRequest(c)
	if err := pr.GetAll(path, &entries); err != nil {
		return nil, fmt.Errorf("failed to fetch schedule entries: %w", err)
	}

	return entries, nil
}

// GetScheduleEntriesInRange fetches entries in a schedule within a date range
func (c *Client) GetScheduleEntriesInRange(ctx context.Context, projectID string, scheduleID int64, startDate, endDate string) ([]ScheduleEntry, error) {
	var entries []ScheduleEntry

	// Build query parameters
	params := url.Values{}
	if startDate != "" {
		params.Set("start_date", startDate)
	}
	if endDate != "" {
		params.Set("end_date", endDate)
	}

	path := fmt.Sprintf("/buckets/%s/schedules/%d/entries.json", projectID, scheduleID)
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	pr := NewPaginatedRequest(c)
	if err := pr.GetAll(path, &entries); err != nil {
		return nil, fmt.Errorf("failed to fetch schedule entries: %w", err)
	}

	return entries, nil
}

// GetUpcomingScheduleEntries fetches upcoming entries in a schedule
func (c *Client) GetUpcomingScheduleEntries(ctx context.Context, projectID string, scheduleID int64) ([]ScheduleEntry, error) {
	var entries []ScheduleEntry
	path := fmt.Sprintf("/buckets/%s/schedules/%d/entries.json?status=upcoming", projectID, scheduleID)

	pr := NewPaginatedRequest(c)
	if err := pr.GetAll(path, &entries); err != nil {
		return nil, fmt.Errorf("failed to fetch upcoming schedule entries: %w", err)
	}

	return entries, nil
}

// GetPastScheduleEntries fetches past entries in a schedule
func (c *Client) GetPastScheduleEntries(ctx context.Context, projectID string, scheduleID int64) ([]ScheduleEntry, error) {
	var entries []ScheduleEntry
	path := fmt.Sprintf("/buckets/%s/schedules/%d/entries.json?status=past", projectID, scheduleID)

	pr := NewPaginatedRequest(c)
	if err := pr.GetAll(path, &entries); err != nil {
		return nil, fmt.Errorf("failed to fetch past schedule entries: %w", err)
	}

	return entries, nil
}

// GetScheduleEntry fetches a specific schedule entry by ID
func (c *Client) GetScheduleEntry(ctx context.Context, projectID string, entryID int64) (*ScheduleEntry, error) {
	var entry ScheduleEntry
	path := fmt.Sprintf("/buckets/%s/schedule_entries/%d.json", projectID, entryID)

	if err := c.Get(path, &entry); err != nil {
		return nil, fmt.Errorf("failed to fetch schedule entry: %w", err)
	}

	return &entry, nil
}

// CreateScheduleEntry creates a new entry in a schedule
func (c *Client) CreateScheduleEntry(ctx context.Context, projectID string, scheduleID int64, req ScheduleEntryCreateRequest) (*ScheduleEntry, error) {
	var entry ScheduleEntry
	path := fmt.Sprintf("/buckets/%s/schedules/%d/entries.json", projectID, scheduleID)

	if err := c.Post(path, req, &entry); err != nil {
		return nil, fmt.Errorf("failed to create schedule entry: %w", err)
	}

	return &entry, nil
}

// UpdateScheduleEntry updates an existing schedule entry
func (c *Client) UpdateScheduleEntry(ctx context.Context, projectID string, entryID int64, req ScheduleEntryUpdateRequest) (*ScheduleEntry, error) {
	var entry ScheduleEntry
	path := fmt.Sprintf("/buckets/%s/schedule_entries/%d.json", projectID, entryID)

	if err := c.Put(path, req, &entry); err != nil {
		return nil, fmt.Errorf("failed to update schedule entry: %w", err)
	}

	return &entry, nil
}

// DeleteScheduleEntry deletes (trashes) a schedule entry
func (c *Client) DeleteScheduleEntry(ctx context.Context, projectID string, entryID int64) error {
	path := fmt.Sprintf("/buckets/%s/recordings/%d/status/trashed.json", projectID, entryID)

	if err := c.Put(path, nil, nil); err != nil {
		return fmt.Errorf("failed to delete schedule entry: %w", err)
	}

	return nil
}
