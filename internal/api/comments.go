package api

import (
	"context"
	"fmt"
	"time"
)

// Comment represents a Basecamp comment
type Comment struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Creator   Person    `json:"creator"`
	Parent    struct {
		ID    int64  `json:"id"`
		Title string `json:"title"`
		Type  string `json:"type"`
		URL   string `json:"url"`
	} `json:"parent"`
	URL string `json:"url"`
}

// CommentCreateRequest represents the payload for creating a comment
type CommentCreateRequest struct {
	Content string `json:"content"`
}

// CommentUpdateRequest represents the payload for updating a comment
type CommentUpdateRequest struct {
	Content string `json:"content"`
}

// ListComments returns all comments for a recording
func (c *Client) ListComments(ctx context.Context, projectID string, recordingID int64) ([]Comment, error) {
	var comments []Comment
	path := fmt.Sprintf("/buckets/%s/recordings/%d/comments.json", projectID, recordingID)

	// Use paginated request to get all comments
	pr := NewPaginatedRequest(c)
	if err := pr.GetAll(path, &comments); err != nil {
		return nil, fmt.Errorf("failed to list comments: %w", err)
	}

	return comments, nil
}

// GetComment returns a specific comment
func (c *Client) GetComment(ctx context.Context, projectID string, commentID int64) (*Comment, error) {
	var comment Comment
	path := fmt.Sprintf("/buckets/%s/comments/%d.json", projectID, commentID)

	if err := c.Get(path, &comment); err != nil {
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	return &comment, nil
}

// CreateComment creates a new comment on a recording
func (c *Client) CreateComment(ctx context.Context, projectID string, recordingID int64, req CommentCreateRequest) (*Comment, error) {
	var comment Comment
	path := fmt.Sprintf("/buckets/%s/recordings/%d/comments.json", projectID, recordingID)

	if err := c.Post(path, req, &comment); err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	return &comment, nil
}

// UpdateComment updates an existing comment
func (c *Client) UpdateComment(ctx context.Context, projectID string, commentID int64, req CommentUpdateRequest) (*Comment, error) {
	var comment Comment
	path := fmt.Sprintf("/buckets/%s/comments/%d.json", projectID, commentID)

	if err := c.Put(path, req, &comment); err != nil {
		return nil, fmt.Errorf("failed to update comment: %w", err)
	}

	return &comment, nil
}

// TrashComment trashes a comment via the recordings status endpoint
func (c *Client) TrashComment(ctx context.Context, projectID string, commentID int64) error {
	path := fmt.Sprintf("/buckets/%s/recordings/%d/status/trashed.json", projectID, commentID)

	if err := c.Put(path, nil, nil); err != nil {
		return fmt.Errorf("failed to trash comment: %w", err)
	}

	return nil
}
