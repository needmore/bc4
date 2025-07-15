package card

import (
	"context"
	"fmt"

	"github.com/needmore/bc4/internal/api"
	"github.com/needmore/bc4/internal/api/project"
)

// Client is the interface for card table-related API calls
type Client interface {
	GetProjectCardTable(ctx context.Context, projectID string) (*api.CardTable, error)
	GetCardTable(ctx context.Context, projectID string, cardTableID int64) (*api.CardTable, error)
	GetCardsInColumn(ctx context.Context, projectID string, columnID int64) ([]api.Card, error)
	GetCard(ctx context.Context, projectID string, cardID int64) (*api.Card, error)
	CreateCard(ctx context.Context, projectID string, columnID int64, req api.CardCreateRequest) (*api.Card, error)
	UpdateCard(ctx context.Context, projectID string, cardID int64, req api.CardUpdateRequest) (*api.Card, error)
	MoveCard(ctx context.Context, projectID string, cardID int64, columnID int64) error
	ArchiveCard(ctx context.Context, projectID string, cardID int64) error
	
	// Step methods
	CreateStep(ctx context.Context, projectID string, cardID int64, req api.StepCreateRequest) (*api.Step, error)
	UpdateStep(ctx context.Context, projectID string, stepID int64, req api.StepUpdateRequest) (*api.Step, error)
	SetStepCompletion(ctx context.Context, projectID string, stepID int64, completed bool) error
	MoveStep(ctx context.Context, projectID string, cardID int64, stepID int64, position int) error
	DeleteStep(ctx context.Context, projectID string, stepID int64) error
}

// client implements the Client interface
type client struct {
	base    *api.BaseClient
	project project.Client
}

// NewClient returns a new card table client
func NewClient(base *api.BaseClient, projectClient project.Client) Client {
	return &client{
		base:    base,
		project: projectClient,
	}
}

// GetProjectCardTable fetches the card table for a project
func (c *client) GetProjectCardTable(ctx context.Context, projectID string) (*api.CardTable, error) {
	// Get project dock to find the card table
	dock, err := c.project.GetProjectDock(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Find the card table in the dock
	for _, tool := range dock {
		if tool.Name == "kanban_board" {
			// Fetch the full card table details
			return c.GetCardTable(ctx, projectID, tool.ID)
		}
	}

	return nil, fmt.Errorf("card table not found for project")
}

// GetCardTable fetches a card table by ID
func (c *client) GetCardTable(ctx context.Context, projectID string, cardTableID int64) (*api.CardTable, error) {
	var cardTable api.CardTable

	path := fmt.Sprintf("/buckets/%s/card_tables/%d.json", projectID, cardTableID)
	if err := c.base.Get(path, &cardTable); err != nil {
		return nil, fmt.Errorf("failed to fetch card table: %w", err)
	}

	return &cardTable, nil
}

// GetCardsInColumn fetches all cards in a specific column
func (c *client) GetCardsInColumn(ctx context.Context, projectID string, columnID int64) ([]api.Card, error) {
	var cards []api.Card
	path := fmt.Sprintf("/buckets/%s/card_tables/lists/%d/cards.json", projectID, columnID)

	// Use paginated request to get all cards
	pr := api.NewPaginatedRequest(c.base)
	if err := pr.GetAll(path, &cards); err != nil {
		return nil, fmt.Errorf("failed to fetch cards: %w", err)
	}

	return cards, nil
}

// GetCard fetches a single card by ID
func (c *client) GetCard(ctx context.Context, projectID string, cardID int64) (*api.Card, error) {
	var card api.Card

	path := fmt.Sprintf("/buckets/%s/card_tables/cards/%d.json", projectID, cardID)
	if err := c.base.Get(path, &card); err != nil {
		return nil, fmt.Errorf("failed to fetch card: %w", err)
	}

	return &card, nil
}

// CreateCard creates a new card in a column
func (c *client) CreateCard(ctx context.Context, projectID string, columnID int64, req api.CardCreateRequest) (*api.Card, error) {
	var card api.Card

	path := fmt.Sprintf("/buckets/%s/card_tables/lists/%d/cards.json", projectID, columnID)
	if err := c.base.Post(path, req, &card); err != nil {
		return nil, fmt.Errorf("failed to create card: %w", err)
	}

	return &card, nil
}

// UpdateCard updates a card
func (c *client) UpdateCard(ctx context.Context, projectID string, cardID int64, req api.CardUpdateRequest) (*api.Card, error) {
	var card api.Card

	path := fmt.Sprintf("/buckets/%s/card_tables/cards/%d.json", projectID, cardID)
	if err := c.base.Put(path, req, &card); err != nil {
		return nil, fmt.Errorf("failed to update card: %w", err)
	}

	return &card, nil
}

// MoveCard moves a card to a different column
func (c *client) MoveCard(ctx context.Context, projectID string, cardID int64, columnID int64) error {
	path := fmt.Sprintf("/buckets/%s/card_tables/cards/%d/moves.json", projectID, cardID)
	req := struct {
		ColumnID int64 `json:"column_id"`
	}{
		ColumnID: columnID,
	}

	if err := c.base.Post(path, req, nil); err != nil {
		return fmt.Errorf("failed to move card: %w", err)
	}

	return nil
}

// ArchiveCard archives a card
func (c *client) ArchiveCard(ctx context.Context, projectID string, cardID int64) error {
	// Cards are archived by moving them to the archive state
	path := fmt.Sprintf("/buckets/%s/recordings/%d/status/archived.json", projectID, cardID)

	if err := c.base.Put(path, nil, nil); err != nil {
		return fmt.Errorf("failed to archive card: %w", err)
	}

	return nil
}

// CreateStep creates a new step in a card
func (c *client) CreateStep(ctx context.Context, projectID string, cardID int64, req api.StepCreateRequest) (*api.Step, error) {
	var step api.Step

	path := fmt.Sprintf("/buckets/%s/card_tables/cards/%d/steps.json", projectID, cardID)
	if err := c.base.Post(path, req, &step); err != nil {
		return nil, fmt.Errorf("failed to create step: %w", err)
	}

	return &step, nil
}

// UpdateStep updates a step
func (c *client) UpdateStep(ctx context.Context, projectID string, stepID int64, req api.StepUpdateRequest) (*api.Step, error) {
	var step api.Step

	path := fmt.Sprintf("/buckets/%s/card_tables/steps/%d.json", projectID, stepID)
	if err := c.base.Put(path, req, &step); err != nil {
		return nil, fmt.Errorf("failed to update step: %w", err)
	}

	return &step, nil
}

// SetStepCompletion sets the completion status of a step
func (c *client) SetStepCompletion(ctx context.Context, projectID string, stepID int64, completed bool) error {
	path := fmt.Sprintf("/buckets/%s/card_tables/steps/%d/completions.json", projectID, stepID)
	completion := "off"
	if completed {
		completion = "on"
	}
	req := struct {
		Completion string `json:"completion"`
	}{
		Completion: completion,
	}

	if err := c.base.Put(path, req, nil); err != nil {
		return fmt.Errorf("failed to set step completion: %w", err)
	}

	return nil
}

// MoveStep repositions a step within a card
func (c *client) MoveStep(ctx context.Context, projectID string, cardID int64, stepID int64, position int) error {
	path := fmt.Sprintf("/buckets/%s/card_tables/cards/%d/positions.json", projectID, cardID)
	req := struct {
		SourceID int64 `json:"source_id"`
		Position int   `json:"position"`
	}{
		SourceID: stepID,
		Position: position,
	}

	if err := c.base.Post(path, req, nil); err != nil {
		return fmt.Errorf("failed to move step: %w", err)
	}

	return nil
}

// DeleteStep deletes a step (archive it)
func (c *client) DeleteStep(ctx context.Context, projectID string, stepID int64) error {
	// Steps are deleted by archiving them
	path := fmt.Sprintf("/buckets/%s/recordings/%d/status/archived.json", projectID, stepID)

	if err := c.base.Put(path, nil, nil); err != nil {
		return fmt.Errorf("failed to delete step: %w", err)
	}

	return nil
}