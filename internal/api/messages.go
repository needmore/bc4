package api

import (
	"context"
	"fmt"
	"time"
)

// MessageBoard represents a Basecamp message board
type MessageBoard struct {
	ID            int64     `json:"id"`
	Title         string    `json:"title"`
	Name          string    `json:"name"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	MessagesURL   string    `json:"messages_url"`
	URL           string    `json:"url"`
	MessagesCount int       `json:"messages_count"`
	Bucket        struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"bucket"`
	Creator Person `json:"creator"`
}

// Message represents a Basecamp message
type Message struct {
	ID        int64     `json:"id"`
	Subject   string    `json:"subject"`
	Content   string    `json:"content"`
	Status    string    `json:"status"`
	Pinned    bool      `json:"pinned"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Creator   Person    `json:"creator"`
	Parent    struct {
		ID    int64  `json:"id"`
		Title string `json:"title"`
		Type  string `json:"type"`
		URL   string `json:"url"`
	} `json:"parent"`
	Category      *MessageCategory `json:"category"`
	CommentsCount int              `json:"comments_count"`
	URL           string           `json:"url"`
}

// MessageCategory represents a message category
type MessageCategory struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

// MessageCreateRequest represents the payload for creating a new message
type MessageCreateRequest struct {
	Subject    string `json:"subject"`
	Content    string `json:"content"`
	Status     string `json:"status,omitempty"` // must be "active" - Basecamp API does not support draft messages
	CategoryID *int64 `json:"category_id,omitempty"`
}

// MessageUpdateRequest represents the payload for updating a message
type MessageUpdateRequest struct {
	Subject    string `json:"subject,omitempty"`
	Content    string `json:"content,omitempty"`
	CategoryID *int64 `json:"category_id,omitempty"`
}

// GetMessageBoard returns the message board for a project
func (c *Client) GetMessageBoard(ctx context.Context, projectID string) (*MessageBoard, error) {
	// First get the project to find its message board
	project, err := c.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Get project tools/features
	path := fmt.Sprintf("/projects/%d.json", project.ID)

	var projectData struct {
		Dock []struct {
			ID    int64  `json:"id"`
			Title string `json:"title"`
			Name  string `json:"name"`
			URL   string `json:"url"`
		} `json:"dock"`
	}

	if err := c.Get(path, &projectData); err != nil {
		return nil, fmt.Errorf("failed to fetch project tools: %w", err)
	}

	// Find the message board in the dock
	for _, tool := range projectData.Dock {
		if tool.Name == "message_board" {
			// Get the full message board details
			var board MessageBoard
			boardPath := fmt.Sprintf("/buckets/%s/message_boards/%d.json", projectID, tool.ID)
			if err := c.Get(boardPath, &board); err != nil {
				return nil, fmt.Errorf("failed to get message board: %w", err)
			}
			return &board, nil
		}
	}

	return nil, fmt.Errorf("message board not found for project")
}

// ListMessages returns all messages on a message board
func (c *Client) ListMessages(ctx context.Context, projectID string, messageBoardID int64) ([]Message, error) {
	var messages []Message
	path := fmt.Sprintf("/buckets/%s/message_boards/%d/messages.json", projectID, messageBoardID)

	// Use paginated request to get all messages
	pr := NewPaginatedRequest(c)
	if err := pr.GetAll(path, &messages); err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	return messages, nil
}

// GetMessage returns a specific message
func (c *Client) GetMessage(ctx context.Context, projectID string, messageID int64) (*Message, error) {
	var message Message
	path := fmt.Sprintf("/buckets/%s/messages/%d.json", projectID, messageID)

	if err := c.Get(path, &message); err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return &message, nil
}

// CreateMessage creates a new message on a message board
func (c *Client) CreateMessage(ctx context.Context, projectID string, messageBoardID int64, req MessageCreateRequest) (*Message, error) {
	var message Message
	path := fmt.Sprintf("/buckets/%s/message_boards/%d/messages.json", projectID, messageBoardID)

	if err := c.Post(path, req, &message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	return &message, nil
}

// UpdateMessage updates an existing message
func (c *Client) UpdateMessage(ctx context.Context, projectID string, messageID int64, req MessageUpdateRequest) (*Message, error) {
	var message Message
	path := fmt.Sprintf("/buckets/%s/messages/%d.json", projectID, messageID)

	if err := c.Put(path, req, &message); err != nil {
		return nil, fmt.Errorf("failed to update message: %w", err)
	}

	return &message, nil
}

// DeleteMessage deletes a message
func (c *Client) DeleteMessage(ctx context.Context, projectID string, messageID int64) error {
	path := fmt.Sprintf("/buckets/%s/messages/%d.json", projectID, messageID)

	if err := c.Delete(path); err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

// ListMessageCategories returns all categories for a message board
func (c *Client) ListMessageCategories(ctx context.Context, projectID string, messageBoardID int64) ([]MessageCategory, error) {
	var categories []MessageCategory
	path := fmt.Sprintf("/buckets/%s/categories.json?categorizable_type=Message::Board&categorizable_id=%d", projectID, messageBoardID)

	// Use paginated request to ensure we get all categories
	pr := NewPaginatedRequest(c)
	if err := pr.GetAll(path, &categories); err != nil {
		return nil, fmt.Errorf("failed to list message categories: %w", err)
	}

	return categories, nil
}

// PinMessage pins a message to the top of the message board
func (c *Client) PinMessage(ctx context.Context, projectID string, messageID int64) error {
	path := fmt.Sprintf("/buckets/%s/recordings/%d/pin.json", projectID, messageID)

	if err := c.Post(path, nil, nil); err != nil {
		return fmt.Errorf("failed to pin message: %w", err)
	}

	return nil
}

// UnpinMessage unpins a message from the top of the message board
func (c *Client) UnpinMessage(ctx context.Context, projectID string, messageID int64) error {
	path := fmt.Sprintf("/buckets/%s/recordings/%d/pin.json", projectID, messageID)

	if err := c.Delete(path); err != nil {
		return fmt.Errorf("failed to unpin message: %w", err)
	}

	return nil
}
