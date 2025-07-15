package project

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/needmore/bc4/internal/api"
)

// Client is the interface for project-related API calls
type Client interface {
	GetProjects(ctx context.Context) ([]api.Project, error)
	GetProject(ctx context.Context, projectID string) (*api.Project, error)
}

// client implements the Client interface
type client struct {
	base *api.BaseClient
}

// NewClient returns a new project client
func NewClient(base *api.BaseClient) Client {
	return &client{base: base}
}

// GetProjects fetches all projects for the account (handles pagination)
func (c *client) GetProjects(ctx context.Context) ([]api.Project, error) {
	var projects []api.Project

	// Use paginated request to get all projects
	pr := api.NewPaginatedRequest(c.base)
	if err := pr.GetAll("/projects.json", &projects); err != nil {
		return nil, fmt.Errorf("failed to fetch projects: %w", err)
	}

	return projects, nil
}

// GetProject fetches a single project by ID
func (c *client) GetProject(ctx context.Context, projectID string) (*api.Project, error) {
	var project api.Project

	path := fmt.Sprintf("/projects/%s.json", projectID)
	if err := c.base.Get(path, &project); err != nil {
		return nil, fmt.Errorf("failed to fetch project: %w", err)
	}

	return &project, nil
}

// GetProjectDock fetches the project dock (tools/features)
func (c *client) GetProjectDock(ctx context.Context, projectID string) ([]api.DockItem, error) {
	project, err := c.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	var projectData struct {
		Dock []api.DockItem `json:"dock"`
	}

	path := fmt.Sprintf("/projects/%d.json", project.ID)
	if err := c.base.Get(path, &projectData); err != nil {
		return nil, fmt.Errorf("failed to fetch project dock: %w", err)
	}

	return projectData.Dock, nil
}