package api

import (
	"context"
	"fmt"
	"time"
)

// Vault represents a Basecamp document vault
type Vault struct {
	ID             int64     `json:"id"`
	Title          string    `json:"title"`
	Name           string    `json:"name"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	DocumentsURL   string    `json:"documents_url"`
	URL            string    `json:"url"`
	DocumentsCount int       `json:"documents_count"`
	Bucket         struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"bucket"`
	Creator Person `json:"creator"`
}

// Document represents a Basecamp document
type Document struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
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
	CommentsCount    int    `json:"comments_count"`
	VisibleToClients bool   `json:"visible_to_clients"`
	URL              string `json:"url"`
}

// DocumentCreateRequest represents the payload for creating a new document
type DocumentCreateRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Status  string `json:"status,omitempty"` // draft or active
}

// DocumentUpdateRequest represents the payload for updating a document
type DocumentUpdateRequest struct {
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
}

// GetVault returns the document vault for a project
func (c *Client) GetVault(ctx context.Context, projectID string) (*Vault, error) {
	// First get the project to find its vault
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

	// Find the vault in the dock
	for _, tool := range projectData.Dock {
		if tool.Name == "vault" {
			// Get the full vault details
			var vault Vault
			vaultPath := fmt.Sprintf("/buckets/%s/vaults/%d.json", projectID, tool.ID)
			if err := c.Get(vaultPath, &vault); err != nil {
				return nil, fmt.Errorf("failed to get vault: %w", err)
			}
			return &vault, nil
		}
	}

	return nil, fmt.Errorf("document vault not found for project")
}

// ListDocuments returns all documents in a vault
func (c *Client) ListDocuments(ctx context.Context, projectID string, vaultID int64) ([]Document, error) {
	var documents []Document
	path := fmt.Sprintf("/buckets/%s/vaults/%d/documents.json", projectID, vaultID)

	// Use paginated request to get all documents
	pr := NewPaginatedRequest(c)
	if err := pr.GetAll(path, &documents); err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	return documents, nil
}

// GetDocument returns a specific document
func (c *Client) GetDocument(ctx context.Context, projectID string, documentID int64) (*Document, error) {
	var document Document
	path := fmt.Sprintf("/buckets/%s/documents/%d.json", projectID, documentID)

	if err := c.Get(path, &document); err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	return &document, nil
}

// CreateDocument creates a new document in a vault
func (c *Client) CreateDocument(ctx context.Context, projectID string, vaultID int64, req DocumentCreateRequest) (*Document, error) {
	var document Document
	path := fmt.Sprintf("/buckets/%s/vaults/%d/documents.json", projectID, vaultID)

	if err := c.Post(path, req, &document); err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	return &document, nil
}

// UpdateDocument updates an existing document
func (c *Client) UpdateDocument(ctx context.Context, projectID string, documentID int64, req DocumentUpdateRequest) (*Document, error) {
	var document Document
	path := fmt.Sprintf("/buckets/%s/documents/%d.json", projectID, documentID)

	if err := c.Put(path, req, &document); err != nil {
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	return &document, nil
}

// DeleteDocument deletes a document
func (c *Client) DeleteDocument(ctx context.Context, projectID string, documentID int64) error {
	path := fmt.Sprintf("/buckets/%s/documents/%d.json", projectID, documentID)

	if err := c.Delete(path); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}
