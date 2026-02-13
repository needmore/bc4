package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

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
	ID         int64        `json:"id"`
	Title      string       `json:"title"`
	Name       string       `json:"name"`
	Type       string       `json:"type"`
	Color      string       `json:"color,omitempty"`
	Status     string       `json:"status"`
	OnHold     OnHoldStatus `json:"on_hold"`
	CardsCount int          `json:"cards_count"`
	CardsURL   string       `json:"cards_url"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
}

// OnHoldStatus represents the on_hold status of a column
type OnHoldStatus struct {
	ID         int64  `json:"id,omitempty"`
	Enabled    bool   `json:"enabled"`
	CardsCount int    `json:"cards_count,omitempty"`
	CardsURL   string `json:"cards_url,omitempty"`
}

// Card represents a card in a card table
type Card struct {
	ID            int64     `json:"id"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	Status        string    `json:"status"`
	DueOn         *string   `json:"due_on,omitempty"`
	Assignees     []Person  `json:"assignees"`
	Steps         []Step    `json:"steps"`
	StepsCount    int       `json:"steps_count"`
	CommentsCount int       `json:"comments_count"`
	Creator       *Person   `json:"creator"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Parent        *Column   `json:"parent"`
	URL           string    `json:"url"`
	IsOnHold      bool      `json:"-"` // Set programmatically, not from API
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

// CardMoveRequest represents the payload for moving a card
type CardMoveRequest struct {
	ColumnID int64 `json:"column_id"`
}

// ColumnCreateRequest represents the payload for creating a new column
type ColumnCreateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
}

// ColumnUpdateRequest represents the payload for updating a column
type ColumnUpdateRequest struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

// ColumnColorRequest represents the payload for changing a column's color
type ColumnColorRequest struct {
	Color string `json:"color"`
}

// ColumnMoveRequest represents the payload for moving a column
type ColumnMoveRequest struct {
	SourceID int64  `json:"source_id"`
	TargetID int64  `json:"target_id"`
	Position string `json:"position,omitempty"` // "before" or "after"
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

// StepCompletionRequest represents the payload for changing step completion
type StepCompletionRequest struct {
	Completion string `json:"completion"` // "on" or "off"
}

// StepPositionRequest represents the payload for repositioning a step
type StepPositionRequest struct {
	SourceID int64 `json:"source_id"`
	Position int   `json:"position"` // zero-indexed
}

// GetAllProjectCardTables fetches all card tables for a project
func (c *Client) GetAllProjectCardTables(ctx context.Context, projectID string) ([]*CardTable, error) {
	// First get the project to find its card tables
	project, err := c.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Get project tools/features
	path := fmt.Sprintf("/projects/%d.json", project.ID)

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch project tools: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var projectData struct {
		Dock []struct {
			ID    int64  `json:"id"`
			Title string `json:"title"`
			Name  string `json:"name"`
			URL   string `json:"url"`
		} `json:"dock"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&projectData); err != nil {
		return nil, fmt.Errorf("failed to decode project data: %w", err)
	}

	// Find all card tables in the dock
	var cardTables []*CardTable
	for _, tool := range projectData.Dock {
		if tool.Name == "kanban_board" {
			// Fetch the full card table details
			cardTable, err := c.GetCardTable(ctx, projectID, tool.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch card table %d: %w", tool.ID, err)
			}
			cardTables = append(cardTables, cardTable)
		}
	}

	if len(cardTables) == 0 {
		return nil, fmt.Errorf("no card tables found for project")
	}

	return cardTables, nil
}

// GetProjectCardTable fetches the default (first) card table for a project
func (c *Client) GetProjectCardTable(ctx context.Context, projectID string) (*CardTable, error) {
	cardTables, err := c.GetAllProjectCardTables(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if len(cardTables) == 0 {
		return nil, fmt.Errorf("no card tables found for project")
	}
	return cardTables[0], nil
}

// GetCardTable fetches a card table by ID
func (c *Client) GetCardTable(ctx context.Context, projectID string, cardTableID int64) (*CardTable, error) {
	var cardTable CardTable

	path := fmt.Sprintf("/buckets/%s/card_tables/%d.json", projectID, cardTableID)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch card table: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err := json.NewDecoder(resp.Body).Decode(&cardTable); err != nil {
		return nil, fmt.Errorf("failed to decode card table: %w", err)
	}

	return &cardTable, nil
}

// GetCardsInColumn fetches all cards in a specific column
func (c *Client) GetCardsInColumn(ctx context.Context, projectID string, columnID int64) ([]Card, error) {
	var cards []Card
	path := fmt.Sprintf("/buckets/%s/card_tables/lists/%d/cards.json", projectID, columnID)

	// Use paginated request to get all cards
	pr := NewPaginatedRequest(c)
	if err := pr.GetAll(path, &cards); err != nil {
		return nil, fmt.Errorf("failed to fetch cards: %w", err)
	}

	return cards, nil
}

// GetOnHoldCardsInColumn fetches all on-hold cards in a column using the on_hold.cards_url
func (c *Client) GetOnHoldCardsInColumn(ctx context.Context, onHoldCardsURL string) ([]Card, error) {
	if onHoldCardsURL == "" {
		return nil, nil
	}

	path := extractPathFromURL(onHoldCardsURL)
	if path == "" {
		return nil, fmt.Errorf("failed to extract path from on-hold cards URL: %s", onHoldCardsURL)
	}

	var cards []Card
	pr := NewPaginatedRequest(c)
	if err := pr.GetAll(path, &cards); err != nil {
		return nil, fmt.Errorf("failed to fetch on-hold cards: %w", err)
	}
	return cards, nil
}

// GetCard fetches a single card by ID
func (c *Client) GetCard(ctx context.Context, projectID string, cardID int64) (*Card, error) {
	var card Card

	path := fmt.Sprintf("/buckets/%s/card_tables/cards/%d.json", projectID, cardID)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch card: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
		return nil, fmt.Errorf("failed to decode card: %w", err)
	}

	return &card, nil
}

// CreateCard creates a new card in a column
func (c *Client) CreateCard(ctx context.Context, projectID string, columnID int64, req CardCreateRequest) (*Card, error) {
	var card Card

	path := fmt.Sprintf("/buckets/%s/card_tables/lists/%d/cards.json", projectID, columnID)
	if err := c.Post(path, req, &card); err != nil {
		return nil, fmt.Errorf("failed to create card: %w", err)
	}

	return &card, nil
}

// UpdateCard updates a card
func (c *Client) UpdateCard(ctx context.Context, projectID string, cardID int64, req CardUpdateRequest) (*Card, error) {
	var card Card

	path := fmt.Sprintf("/buckets/%s/card_tables/cards/%d.json", projectID, cardID)
	if err := c.Put(path, req, &card); err != nil {
		return nil, fmt.Errorf("failed to update card: %w", err)
	}

	return &card, nil
}

// MoveCard moves a card to a different column
func (c *Client) MoveCard(ctx context.Context, projectID string, cardID int64, columnID int64) error {
	path := fmt.Sprintf("/buckets/%s/card_tables/cards/%d/moves.json", projectID, cardID)
	req := CardMoveRequest{ColumnID: columnID}

	if err := c.Post(path, req, nil); err != nil {
		return fmt.Errorf("failed to move card: %w", err)
	}

	return nil
}

// ArchiveCard archives a card
func (c *Client) ArchiveCard(ctx context.Context, projectID string, cardID int64) error {
	// Cards are archived by moving them to the archive state
	path := fmt.Sprintf("/buckets/%s/recordings/%d/status/archived.json", projectID, cardID)

	if err := c.Put(path, nil, nil); err != nil {
		return fmt.Errorf("failed to archive card: %w", err)
	}

	return nil
}

// CreateColumn creates a new column in a card table
func (c *Client) CreateColumn(ctx context.Context, projectID string, cardTableID int64, req ColumnCreateRequest) (*Column, error) {
	var column Column

	path := fmt.Sprintf("/buckets/%s/card_tables/%d/columns.json", projectID, cardTableID)
	if err := c.Post(path, req, &column); err != nil {
		return nil, fmt.Errorf("failed to create column: %w", err)
	}

	return &column, nil
}

// UpdateColumn updates a column
func (c *Client) UpdateColumn(ctx context.Context, projectID string, columnID int64, req ColumnUpdateRequest) (*Column, error) {
	var column Column

	path := fmt.Sprintf("/buckets/%s/card_tables/columns/%d.json", projectID, columnID)
	if err := c.Put(path, req, &column); err != nil {
		return nil, fmt.Errorf("failed to update column: %w", err)
	}

	return &column, nil
}

// SetColumnColor sets the color of a column
func (c *Client) SetColumnColor(ctx context.Context, projectID string, columnID int64, color string) error {
	path := fmt.Sprintf("/buckets/%s/card_tables/columns/%d/color.json", projectID, columnID)
	req := ColumnColorRequest{Color: color}

	if err := c.Put(path, req, nil); err != nil {
		return fmt.Errorf("failed to set column color: %w", err)
	}

	return nil
}

// MoveColumn moves a column to a different position
func (c *Client) MoveColumn(ctx context.Context, projectID string, cardTableID int64, sourceID, targetID int64, position string) error {
	path := fmt.Sprintf("/buckets/%s/card_tables/%d/moves.json", projectID, cardTableID)
	req := ColumnMoveRequest{
		SourceID: sourceID,
		TargetID: targetID,
		Position: position,
	}

	if err := c.Post(path, req, nil); err != nil {
		return fmt.Errorf("failed to move column: %w", err)
	}

	return nil
}

// CreateStep creates a new step in a card
func (c *Client) CreateStep(ctx context.Context, projectID string, cardID int64, req StepCreateRequest) (*Step, error) {
	var step Step

	path := fmt.Sprintf("/buckets/%s/card_tables/cards/%d/steps.json", projectID, cardID)
	if err := c.Post(path, req, &step); err != nil {
		return nil, fmt.Errorf("failed to create step: %w", err)
	}

	return &step, nil
}

// UpdateStep updates a step
func (c *Client) UpdateStep(ctx context.Context, projectID string, stepID int64, req StepUpdateRequest) (*Step, error) {
	var step Step

	path := fmt.Sprintf("/buckets/%s/card_tables/steps/%d.json", projectID, stepID)
	if err := c.Put(path, req, &step); err != nil {
		return nil, fmt.Errorf("failed to update step: %w", err)
	}

	return &step, nil
}

// SetStepCompletion sets the completion status of a step
func (c *Client) SetStepCompletion(ctx context.Context, projectID string, stepID int64, completed bool) error {
	path := fmt.Sprintf("/buckets/%s/card_tables/steps/%d/completions.json", projectID, stepID)
	completion := "off"
	if completed {
		completion = "on"
	}
	req := StepCompletionRequest{Completion: completion}

	if err := c.Put(path, req, nil); err != nil {
		return fmt.Errorf("failed to set step completion: %w", err)
	}

	return nil
}

// MoveStep repositions a step within a card
func (c *Client) MoveStep(ctx context.Context, projectID string, cardID int64, stepID int64, position int) error {
	path := fmt.Sprintf("/buckets/%s/card_tables/cards/%d/positions.json", projectID, cardID)
	req := StepPositionRequest{
		SourceID: stepID,
		Position: position,
	}

	if err := c.Post(path, req, nil); err != nil {
		return fmt.Errorf("failed to move step: %w", err)
	}

	return nil
}

// DeleteStep deletes a step (archive it)
func (c *Client) DeleteStep(ctx context.Context, projectID string, stepID int64) error {
	// Steps are deleted by archiving them
	path := fmt.Sprintf("/buckets/%s/recordings/%d/status/archived.json", projectID, stepID)

	if err := c.Put(path, nil, nil); err != nil {
		return fmt.Errorf("failed to delete step: %w", err)
	}

	return nil
}

// SetColumnOnHold marks a column as on-hold
func (c *Client) SetColumnOnHold(ctx context.Context, projectID string, columnID int64) error {
	path := fmt.Sprintf("/buckets/%s/card_tables/columns/%d/on_hold.json", projectID, columnID)

	if err := c.Post(path, nil, nil); err != nil {
		return fmt.Errorf("failed to set column on-hold: %w", err)
	}

	return nil
}

// RemoveColumnOnHold removes the on-hold status from a column
func (c *Client) RemoveColumnOnHold(ctx context.Context, projectID string, columnID int64) error {
	path := fmt.Sprintf("/buckets/%s/card_tables/columns/%d/on_hold.json", projectID, columnID)

	if err := c.Delete(path); err != nil {
		return fmt.Errorf("failed to remove column on-hold: %w", err)
	}

	return nil
}
