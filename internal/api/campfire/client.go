package campfire

import (
	"context"
	"fmt"
	"strings"

	"github.com/needmore/bc4/internal/api"
)

// Client is the interface for campfire-related API calls
type Client interface {
	ListCampfires(ctx context.Context, projectID string) ([]api.Campfire, error)
	GetCampfire(ctx context.Context, projectID string, campfireID int64) (*api.Campfire, error)
	GetCampfireByName(ctx context.Context, projectID string, name string) (*api.Campfire, error)
	GetCampfireLines(ctx context.Context, projectID string, campfireID int64, limit int) ([]api.CampfireLine, error)
	PostCampfireLine(ctx context.Context, projectID string, campfireID int64, content string) (*api.CampfireLine, error)
	DeleteCampfireLine(ctx context.Context, projectID string, campfireID int64, lineID int64) error
}

// client implements the Client interface
type client struct {
	base *api.BaseClient
}

// NewClient returns a new campfire client
func NewClient(base *api.BaseClient) Client {
	return &client{base: base}
}

// ListCampfires returns all campfires for a project
func (c *client) ListCampfires(ctx context.Context, projectID string) ([]api.Campfire, error) {
	var campfires []api.Campfire
	path := fmt.Sprintf("/buckets/%s/chats.json", projectID)

	// Use paginated request to get all campfires
	pr := api.NewPaginatedRequest(c.base)
	if err := pr.GetAll(path, &campfires); err != nil {
		return nil, fmt.Errorf("failed to list campfires: %w", err)
	}

	return campfires, nil
}

// GetCampfire returns a specific campfire
func (c *client) GetCampfire(ctx context.Context, projectID string, campfireID int64) (*api.Campfire, error) {
	var campfire api.Campfire
	path := fmt.Sprintf("/buckets/%s/chats/%d.json", projectID, campfireID)

	if err := c.base.Get(path, &campfire); err != nil {
		return nil, fmt.Errorf("failed to get campfire: %w", err)
	}

	return &campfire, nil
}

// GetCampfireLines returns messages from a campfire
func (c *client) GetCampfireLines(ctx context.Context, projectID string, campfireID int64, limit int) ([]api.CampfireLine, error) {
	var lines []api.CampfireLine
	path := fmt.Sprintf("/buckets/%s/chats/%d/lines.json", projectID, campfireID)

	// If limit is specified, just get one page with that limit
	if limit > 0 {
		path = fmt.Sprintf("%s?limit=%d", path, limit)
		if err := c.base.Get(path, &lines); err != nil {
			return nil, fmt.Errorf("failed to get campfire lines: %w", err)
		}
		return lines, nil
	}

	// Otherwise, use paginated request to get all lines
	pr := api.NewPaginatedRequest(c.base)
	if err := pr.GetAll(path, &lines); err != nil {
		return nil, fmt.Errorf("failed to get campfire lines: %w", err)
	}

	return lines, nil
}

// PostCampfireLine posts a new message to a campfire
func (c *client) PostCampfireLine(ctx context.Context, projectID string, campfireID int64, content string) (*api.CampfireLine, error) {
	var line api.CampfireLine
	path := fmt.Sprintf("/buckets/%s/chats/%d/lines.json", projectID, campfireID)

	payload := api.CampfireLineCreate{
		Content: content,
	}

	if err := c.base.Post(path, payload, &line); err != nil {
		return nil, fmt.Errorf("failed to post campfire line: %w", err)
	}

	return &line, nil
}

// DeleteCampfireLine deletes a message from a campfire
func (c *client) DeleteCampfireLine(ctx context.Context, projectID string, campfireID int64, lineID int64) error {
	path := fmt.Sprintf("/buckets/%s/chats/%d/lines/%d.json", projectID, campfireID, lineID)

	if err := c.base.Delete(path); err != nil {
		return fmt.Errorf("failed to delete campfire line: %w", err)
	}

	return nil
}

// GetCampfireByName finds a campfire by name (case-insensitive partial match)
func (c *client) GetCampfireByName(ctx context.Context, projectID string, name string) (*api.Campfire, error) {
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