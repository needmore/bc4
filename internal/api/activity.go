package api

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// Event represents a Basecamp activity event
type Event struct {
	ID            int64     `json:"id"`
	Action        string    `json:"action"`
	CreatedAt     time.Time `json:"created_at"`
	RecordingType string    `json:"recording_type"`
	Recording     Recording `json:"recording"`
	Creator       Person    `json:"creator"`
	Bucket        Bucket    `json:"bucket"`
}

// Recording represents a Basecamp recording (generic content item)
type Recording struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	URL       string    `json:"url"`
	AppURL    string    `json:"app_url"`
	Creator   Person    `json:"creator"`
	Bucket    Bucket    `json:"bucket"`
	Parent    *Parent   `json:"parent,omitempty"`
}

// Bucket represents a Basecamp bucket (project container)
type Bucket struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// Parent represents a parent recording
type Parent struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	Type   string `json:"type"`
	URL    string `json:"url"`
	AppURL string `json:"app_url"`
}

// ActivityListOptions contains options for listing activity
type ActivityListOptions struct {
	Since          *time.Time // Filter events since this time
	RecordingTypes []string   // Filter by recording types (todo, message, document, etc.)
	Limit          int        // Maximum number of events to return
}

// ListEvents returns activity events for a recording
func (c *Client) ListEvents(ctx context.Context, projectID string, recordingID int64) ([]Event, error) {
	var events []Event
	path := fmt.Sprintf("/buckets/%s/recordings/%d/events.json", projectID, recordingID)

	pr := NewPaginatedRequest(c)
	if err := pr.GetAll(path, &events); err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	return events, nil
}

// ListRecordings returns all recordings (activity items) for a project
func (c *Client) ListRecordings(ctx context.Context, projectID string, opts *ActivityListOptions) ([]Recording, error) {
	// Default types to fetch if none specified
	typesToFetch := []string{"Todo", "Message", "Document", "Comment"}

	if opts != nil && len(opts.RecordingTypes) > 0 {
		typesToFetch = opts.RecordingTypes
	}

	var allRecordings []Recording

	// Fetch recordings for each type
	for _, recordingType := range typesToFetch {
		typeRecordings, err := c.listRecordingsByType(ctx, projectID, recordingType, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list %s recordings: %w", recordingType, err)
		}
		allRecordings = append(allRecordings, typeRecordings...)
	}

	// Sort by updated_at descending
	sortRecordings(allRecordings)

	// Apply filtering if options provided
	if opts != nil {
		allRecordings = filterRecordings(allRecordings, opts)
	}

	return allRecordings, nil
}

// listRecordingsByType fetches recordings of a specific type for a project
func (c *Client) listRecordingsByType(ctx context.Context, projectID string, recordingType string, opts *ActivityListOptions) ([]Recording, error) {
	var recordings []Recording

	// Build query params
	params := url.Values{}
	params.Set("bucket", projectID)
	params.Set("type", recordingType)
	params.Set("sort", "updated_at")
	params.Set("direction", "desc")

	path := fmt.Sprintf("/projects/recordings.json?%s", params.Encode())

	pr := NewPaginatedRequest(c)
	if err := pr.GetAll(path, &recordings); err != nil {
		return nil, err
	}

	return recordings, nil
}

// sortRecordings sorts recordings by updated_at in descending order
func sortRecordings(recordings []Recording) {
	// Sort by updated_at descending (most recent first)
	for i := 0; i < len(recordings); i++ {
		for j := i + 1; j < len(recordings); j++ {
			if recordings[j].UpdatedAt.After(recordings[i].UpdatedAt) {
				recordings[i], recordings[j] = recordings[j], recordings[i]
			}
		}
	}
}

// filterRecordings applies filtering options to recordings
func filterRecordings(recordings []Recording, opts *ActivityListOptions) []Recording {
	var filtered []Recording

	for _, r := range recordings {
		// Filter by since time
		if opts.Since != nil && r.UpdatedAt.Before(*opts.Since) {
			continue
		}

		filtered = append(filtered, r)

		// Apply limit
		if opts.Limit > 0 && len(filtered) >= opts.Limit {
			break
		}
	}

	return filtered
}

// GetRecording returns a specific recording
func (c *Client) GetRecording(ctx context.Context, projectID string, recordingID int64) (*Recording, error) {
	var recording Recording
	path := fmt.Sprintf("/buckets/%s/recordings/%d.json", projectID, recordingID)

	if err := c.Get(path, &recording); err != nil {
		return nil, fmt.Errorf("failed to get recording: %w", err)
	}

	return &recording, nil
}
