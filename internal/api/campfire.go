package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

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
	Content     string `json:"content"`
	ContentType string `json:"content_type,omitempty"`
}

// ListCampfires returns all campfires for a project
func (c *Client) ListCampfires(ctx context.Context, projectID string) ([]Campfire, error) {
	var campfires []Campfire
	path := fmt.Sprintf("/buckets/%s/chats.json", projectID)

	// Use paginated request to get all campfires
	pr := NewPaginatedRequest(c)
	if err := pr.GetAll(path, &campfires); err != nil {
		return nil, fmt.Errorf("failed to list campfires: %w", err)
	}

	return campfires, nil
}

// GetCampfire returns a specific campfire
func (c *Client) GetCampfire(ctx context.Context, projectID string, campfireID int64) (*Campfire, error) {
	var campfire Campfire
	path := fmt.Sprintf("/buckets/%s/chats/%d.json", projectID, campfireID)

	if err := c.Get(path, &campfire); err != nil {
		return nil, fmt.Errorf("failed to get campfire: %w", err)
	}

	return &campfire, nil
}

// GetCampfireLines returns messages from a campfire
func (c *Client) GetCampfireLines(ctx context.Context, projectID string, campfireID int64, limit int) ([]CampfireLine, error) {
	var lines []CampfireLine
	path := fmt.Sprintf("/buckets/%s/chats/%d/lines.json", projectID, campfireID)

	// If limit is specified, just get one page with that limit
	if limit > 0 {
		path = fmt.Sprintf("%s?limit=%d", path, limit)
		if err := c.Get(path, &lines); err != nil {
			return nil, fmt.Errorf("failed to get campfire lines: %w", err)
		}
		return lines, nil
	}

	// Otherwise, use paginated request to get all lines
	pr := NewPaginatedRequest(c)
	if err := pr.GetAll(path, &lines); err != nil {
		return nil, fmt.Errorf("failed to get campfire lines: %w", err)
	}

	return lines, nil
}

// PostCampfireLine posts a new message to a campfire
func (c *Client) PostCampfireLine(ctx context.Context, projectID string, campfireID int64, content string, contentType string) (*CampfireLine, error) {
	var line CampfireLine
	path := fmt.Sprintf("/buckets/%s/chats/%d/lines.json", projectID, campfireID)

	payload := CampfireLineCreate{
		Content:     content,
		ContentType: contentType,
	}

	if err := c.Post(path, payload, &line); err != nil {
		return nil, fmt.Errorf("failed to post campfire line: %w", err)
	}

	return &line, nil
}

// DeleteCampfireLine deletes a message from a campfire
func (c *Client) DeleteCampfireLine(ctx context.Context, projectID string, campfireID int64, lineID int64) error {
	path := fmt.Sprintf("/buckets/%s/chats/%d/lines/%d.json", projectID, campfireID, lineID)

	if err := c.Delete(path); err != nil {
		return fmt.Errorf("failed to delete campfire line: %w", err)
	}

	return nil
}

// GetCampfireByName finds a campfire by name (case-insensitive partial match)
func (c *Client) GetCampfireByName(ctx context.Context, projectID string, name string) (*Campfire, error) {
	campfires, err := c.ListCampfires(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Try exact match first
	for _, cf := range campfires {
		if cf.Name == name {
			return &cf, nil
		}
	}

	// Try case-insensitive partial match
	for _, cf := range campfires {
		if containsIgnoreCase(cf.Name, name) {
			return &cf, nil
		}
	}

	return nil, fmt.Errorf("campfire not found: %s", name)
}

// containsIgnoreCase checks if a string contains another string (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
