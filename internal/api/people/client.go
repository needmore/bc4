package people

import (
	"context"
	"fmt"

	"github.com/needmore/bc4/internal/api"
)

// Client is the interface for people-related API calls
type Client interface {
	GetProjectPeople(ctx context.Context, projectID string) ([]api.Person, error)
	GetPerson(ctx context.Context, personID int64) (*api.Person, error)
}

// client implements the Client interface
type client struct {
	base *api.BaseClient
}

// NewClient returns a new people client
func NewClient(base *api.BaseClient) Client {
	return &client{base: base}
}

// GetProjectPeople fetches all people associated with a project
func (c *client) GetProjectPeople(ctx context.Context, projectID string) ([]api.Person, error) {
	var people []api.Person
	path := fmt.Sprintf("/projects/%s/people.json", projectID)

	// Use paginated request to get all people
	pr := api.NewPaginatedRequest(c.base)
	if err := pr.GetAll(path, &people); err != nil {
		return nil, fmt.Errorf("failed to fetch project people: %w", err)
	}

	return people, nil
}

// GetPerson fetches a specific person by ID
func (c *client) GetPerson(ctx context.Context, personID int64) (*api.Person, error) {
	var person api.Person

	path := fmt.Sprintf("/people/%d.json", personID)
	if err := c.base.Get(path, &person); err != nil {
		return nil, fmt.Errorf("failed to fetch person: %w", err)
	}

	return &person, nil
}